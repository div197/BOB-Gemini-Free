package gemini

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
)

const (
	maxStreamSubscriberBuffer = 200
	maxStreamHistoryChunks    = 4096
	maxStreamHistoryBytes     = 16 << 20
)

var (
	// ErrStreamSubscriberTooSlow is returned to a follower whose bounded queue
	// filled. Dropping a stream delta would silently corrupt the response, so
	// the follower is disconnected with an explicit error instead.
	ErrStreamSubscriberTooSlow = errors.New("stream subscriber could not keep up; response was not replayed completely")
	ErrStreamHistoryLimit      = errors.New("stream history limit reached; response cannot be replayed to a new subscriber")
	ErrStreamRunnerNil         = errors.New("stream runner is nil")
	ErrStreamEmitterNil        = errors.New("stream emitter is nil")
)

type streamSubscriber struct {
	ch     chan string
	err    error
	closed bool
}

type activeStream struct {
	mu               sync.Mutex
	subscribers      map[*streamSubscriber]struct{}
	history          []string
	historyBytes     int
	historyTruncated bool
	done             bool
	abandoned        bool
	err              error
	ctx              context.Context
	cancel           context.CancelFunc
}

type nonStreamFlight struct {
	done chan struct{}
	val  string
	err  error
}

// StreamFlight multiplexes concurrent identical stream and non-stream requests so that
// multiple students requesting the same prompt/exercise simultaneously
// share a single upstream request without duplicating traffic or getting blocked.
type StreamFlight struct {
	mu         sync.Mutex
	streams    map[string]*activeStream
	nonStreams map[string]*nonStreamFlight
}

func NewStreamFlight() *StreamFlight {
	return &StreamFlight{
		streams:    make(map[string]*activeStream),
		nonStreams: make(map[string]*nonStreamFlight),
	}
}

func (sf *StreamFlight) Execute(key string, runUpstream func() (string, error)) (string, error) {
	return sf.ExecuteContext(context.Background(), key, runUpstream)
}

func (sf *StreamFlight) ExecuteContext(ctx context.Context, key string, runUpstream func() (string, error)) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if runUpstream == nil {
		return "", ErrStreamRunnerNil
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if key == "" {
		return runUpstream()
	}

	sf.mu.Lock()
	if sf.nonStreams == nil {
		sf.nonStreams = make(map[string]*nonStreamFlight)
	}
	flight, exists := sf.nonStreams[key]
	if !exists {
		flight = &nonStreamFlight{done: make(chan struct{})}
		sf.nonStreams[key] = flight
		sf.mu.Unlock()

		defer func() {
			sf.mu.Lock()
			delete(sf.nonStreams, key)
			sf.mu.Unlock()
		}()

		val, err := runUpstream()
		flight.val = val
		flight.err = err
		close(flight.done)
		return val, err
	}
	sf.mu.Unlock()

	select {
	case <-flight.done:
		return flight.val, flight.err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func (sf *StreamFlight) Key(prompt string, modelID, thinkMode int, fileRefs []string) string {
	return sf.KeyWithScope("", prompt, modelID, thinkMode, fileRefs)
}

// KeyWithScope derives a request-coalescing key with an explicit trust scope.
// The scope is intentionally supplied by the caller rather than inferred from
// prompt text. Session-bound callers can either use a stable non-secret scope
// or disable coalescing entirely when the upstream response may be
// personalized.
func (sf *StreamFlight) KeyWithScope(scope, prompt string, modelID, thinkMode int, fileRefs []string) string {
	normalizedPrompt := strings.TrimSpace(prompt)
	h := sha256.New()
	fmt.Fprintf(h, "%s:%d:%d:%s", scope, modelID, thinkMode, normalizedPrompt)
	for _, ref := range fileRefs {
		fmt.Fprintf(h, ":%s", ref)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (sf *StreamFlight) ExecuteStream(key string, runUpstream func(emit func(string) error) error, emit func(string) error) error {
	return sf.ExecuteStreamContext(context.Background(), key, runUpstream, emit)
}

func (sf *StreamFlight) ExecuteStreamContext(ctx context.Context, key string, runUpstream func(emit func(string) error) error, emit func(string) error) error {
	if runUpstream == nil {
		return ErrStreamRunnerNil
	}
	return sf.ExecuteStreamContextWithRunner(ctx, key, func(_ context.Context, streamEmit func(string) error) error {
		return runUpstream(streamEmit)
	}, emit)
}

// ExecuteStreamContextWithRunner multiplexes a stream while giving the
// upstream runner a context owned by the shared flight. A caller cancelling
// its own request detaches only that caller; the shared upstream is cancelled
// only after the last subscriber leaves or the runner completes. This is
// important for coalesced classroom requests: the first HTTP client must not
// be able to terminate another client's otherwise healthy response.
func (sf *StreamFlight) ExecuteStreamContextWithRunner(ctx context.Context, key string, runUpstream func(context.Context, func(string) error) error, emit func(string) error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if runUpstream == nil {
		return ErrStreamRunnerNil
	}
	if emit == nil {
		return ErrStreamEmitterNil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if key == "" {
		return runUpstream(ctx, emit)
	}

	var (
		stream    *activeStream
		history   []string
		streamErr error
		start     bool
	)

	// sf.mu -> stream.mu is the lock order whenever both locks are needed.
	// The loop also handles a stream that was abandoned after its last caller
	// disconnected but before its runner had returned.
	for {
		sf.mu.Lock()
		if sf.streams == nil {
			sf.streams = make(map[string]*activeStream)
		}
		stream = sf.streams[key]
		if stream == nil {
			sharedCtx, cancel := sharedStreamContext(ctx)
			stream = &activeStream{
				subscribers: make(map[*streamSubscriber]struct{}),
				ctx:         sharedCtx,
				cancel:      cancel,
			}
			sf.streams[key] = stream
			start = true
		}

		sub := &streamSubscriber{ch: make(chan string, maxStreamSubscriberBuffer)}
		stream.mu.Lock()
		if stream.abandoned {
			stream.mu.Unlock()
			if sf.streams[key] == stream {
				delete(sf.streams, key)
			}
			sf.mu.Unlock()
			start = false
			continue
		}
		if stream.done {
			if stream.historyTruncated {
				streamErr = stream.err
				stream.mu.Unlock()
				sf.mu.Unlock()
				if streamErr != nil {
					return streamErr
				}
				return ErrStreamHistoryLimit
			}
			history = append([]string(nil), stream.history...)
			streamErr = stream.err
			stream.mu.Unlock()
			sf.mu.Unlock()
			break
		}
		if stream.historyTruncated {
			stream.mu.Unlock()
			sf.mu.Unlock()
			return ErrStreamHistoryLimit
		}

		// Register before replaying. The publisher holds stream.mu while it
		// appends history and queues live deltas, so no boundary is lost.
		history = append([]string(nil), stream.history...)
		stream.subscribers[sub] = struct{}{}
		stream.mu.Unlock()
		sf.mu.Unlock()

		if start {
			go sf.runStream(key, stream, runUpstream)
		}
		return sf.consumeStream(ctx, key, stream, sub, history, emit)
	}

	for _, delta := range history {
		if err := ctx.Err(); err != nil {
			return err
		}
		if emitErr := emit(delta); emitErr != nil {
			return emitErr
		}
	}
	return streamErr
}

func sharedStreamContext(parent context.Context) (context.Context, context.CancelFunc) {
	// A participant's deadline is a property of that HTTP request, not of the
	// coalesced upstream operation. The actual runner owns its own total
	// deadline (the Gemini client uses http.Client.Timeout); inheriting the
	// first participant's deadline would let a short-lived caller terminate a
	// longer-lived follower's shared request.
	return context.WithCancel(context.WithoutCancel(parent))
}

func (sf *StreamFlight) runStream(key string, stream *activeStream, runUpstream func(context.Context, func(string) error) error) {
	err := runUpstream(stream.ctx, func(delta string) error {
		if ctxErr := stream.ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		stream.mu.Lock()
		if !stream.historyTruncated {
			if len(stream.history) >= maxStreamHistoryChunks || stream.historyBytes+len(delta) > maxStreamHistoryBytes {
				stream.historyTruncated = true
			} else {
				stream.history = append(stream.history, delta)
				stream.historyBytes += len(delta)
			}
		}
		for sub := range stream.subscribers {
			select {
			case sub.ch <- delta:
			default:
				closeStreamSubscriberLocked(stream, sub, ErrStreamSubscriberTooSlow)
			}
		}
		stream.mu.Unlock()
		return nil
	})

	stream.mu.Lock()
	stream.done = true
	stream.err = err
	for sub := range stream.subscribers {
		closeStreamSubscriberLocked(stream, sub, nil)
	}
	stream.mu.Unlock()
	stream.cancel()

	// Do not retain completed streams in the global map. A late caller can
	// still replay the completed history during this short hand-off window;
	// callers after deletion start a fresh flight.
	sf.mu.Lock()
	if sf.streams[key] == stream {
		delete(sf.streams, key)
	}
	sf.mu.Unlock()
}

func (sf *StreamFlight) consumeStream(ctx context.Context, key string, stream *activeStream, sub *streamSubscriber, history []string, emit func(string) error) error {
	defer sf.detachStreamSubscriber(key, stream, sub)

	for _, delta := range history {
		if err := ctx.Err(); err != nil {
			return err
		}
		if emitErr := emit(delta); emitErr != nil {
			return emitErr
		}
	}

	for {
		// Prefer a caller cancellation over a concurrently queued delta. A
		// cancelled HTTP request must not report success merely because the
		// upstream published one final item at the same time; the shared
		// flight remains available to its other subscribers.
		if err := ctx.Err(); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case delta, ok := <-sub.ch:
			if err := ctx.Err(); err != nil {
				return err
			}
			if !ok {
				stream.mu.Lock()
				err := sub.err
				if err == nil {
					err = stream.err
				}
				stream.mu.Unlock()
				return err
			}
			if emitErr := emit(delta); emitErr != nil {
				return emitErr
			}
		}
	}
}

func (sf *StreamFlight) detachStreamSubscriber(key string, stream *activeStream, sub *streamSubscriber) {
	// Serialize detachment with new subscriptions so the last caller cannot
	// cancel a flight that a new caller has just joined.
	sf.mu.Lock()
	stream.mu.Lock()
	closeStreamSubscriberLocked(stream, sub, nil)
	if !stream.done && len(stream.subscribers) == 0 {
		stream.abandoned = true
		stream.cancel()
		if sf.streams[key] == stream {
			delete(sf.streams, key)
		}
	}
	stream.mu.Unlock()
	sf.mu.Unlock()
}

func closeStreamSubscriberLocked(stream *activeStream, sub *streamSubscriber, err error) {
	if sub == nil || sub.closed {
		return
	}
	if err != nil {
		sub.err = err
	}
	sub.closed = true
	delete(stream.subscribers, sub)
	close(sub.ch)
}
