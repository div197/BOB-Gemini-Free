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
	err              error
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
		return runUpstream(emit)
	}

	sf.mu.Lock()
	if sf.streams == nil {
		sf.streams = make(map[string]*activeStream)
	}
	stream, exists := sf.streams[key]
	if !exists {
		// Leader request: initiates the upstream stream
		stream = &activeStream{subscribers: make(map[*streamSubscriber]struct{})}
		sf.streams[key] = stream
		sf.mu.Unlock()

		defer func() {
			sf.mu.Lock()
			delete(sf.streams, key)
			sf.mu.Unlock()
		}()

		err := runUpstream(func(delta string) error {
			if ctxErr := ctx.Err(); ctxErr != nil {
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

			return emit(delta)
		})

		stream.mu.Lock()
		stream.done = true
		stream.err = err
		for sub := range stream.subscribers {
			closeStreamSubscriberLocked(stream, sub, nil)
		}
		stream.mu.Unlock()

		return err
	}

	// Follower request: joins existing active stream
	sub := &streamSubscriber{ch: make(chan string, maxStreamSubscriberBuffer)}
	stream.mu.Lock()
	if stream.done {
		// Stream already finished: replay complete history
		if stream.historyTruncated {
			streamErr := stream.err
			stream.mu.Unlock()
			sf.mu.Unlock()
			if streamErr != nil {
				return streamErr
			}
			return ErrStreamHistoryLimit
		}
		history := make([]string, len(stream.history))
		copy(history, stream.history)
		err := stream.err
		stream.mu.Unlock()
		sf.mu.Unlock()

		for _, delta := range history {
			if err := ctx.Err(); err != nil {
				return err
			}
			if emitErr := emit(delta); emitErr != nil {
				return emitErr
			}
		}
		return err
	}
	if stream.historyTruncated {
		stream.mu.Unlock()
		sf.mu.Unlock()
		return ErrStreamHistoryLimit
	}

	// Replay history accumulated so far
	history := make([]string, len(stream.history))
	copy(history, stream.history)
	stream.subscribers[sub] = struct{}{}
	stream.mu.Unlock()
	sf.mu.Unlock()

	defer func() {
		stream.mu.Lock()
		closeStreamSubscriberLocked(stream, sub, nil)
		stream.mu.Unlock()
	}()

	for _, delta := range history {
		if err := ctx.Err(); err != nil {
			return err
		}
		if emitErr := emit(delta); emitErr != nil {
			return emitErr
		}
	}

	// Stream new live deltas as they arrive
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case delta, ok := <-sub.ch:
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
