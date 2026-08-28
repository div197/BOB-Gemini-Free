package gemini

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
)

type subscriber chan string

type activeStream struct {
	mu          sync.Mutex
	subscribers []subscriber
	history     []string
	done        bool
	err         error
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

	<-flight.done
	return flight.val, flight.err
}

func (sf *StreamFlight) Key(prompt string, modelID, thinkMode int, fileRefs []string) string {
	normalizedPrompt := strings.TrimSpace(prompt)
	h := sha256.New()
	fmt.Fprintf(h, "%d:%d:%s", modelID, thinkMode, normalizedPrompt)
	for _, ref := range fileRefs {
		fmt.Fprintf(h, ":%s", ref)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (sf *StreamFlight) ExecuteStream(key string, runUpstream func(emit func(string) error) error, emit func(string) error) error {
	if key == "" {
		return runUpstream(emit)
	}

	sf.mu.Lock()
	stream, exists := sf.streams[key]
	if !exists {
		// Leader request: initiates the upstream stream
		stream = &activeStream{}
		sf.streams[key] = stream
		sf.mu.Unlock()

		defer func() {
			sf.mu.Lock()
			delete(sf.streams, key)
			sf.mu.Unlock()
		}()

		err := runUpstream(func(delta string) error {
			stream.mu.Lock()
			stream.history = append(stream.history, delta)
			for _, sub := range stream.subscribers {
				select {
				case sub <- delta:
				default:
				}
			}
			stream.mu.Unlock()

			return emit(delta)
		})

		stream.mu.Lock()
		stream.done = true
		stream.err = err
		for _, sub := range stream.subscribers {
			close(sub)
		}
		stream.mu.Unlock()

		return err
	}

	// Follower request: joins existing active stream
	sub := make(subscriber, 200)
	stream.mu.Lock()
	if stream.done {
		// Stream already finished: replay complete history
		history := make([]string, len(stream.history))
		copy(history, stream.history)
		err := stream.err
		stream.mu.Unlock()
		sf.mu.Unlock()

		for _, delta := range history {
			if emitErr := emit(delta); emitErr != nil {
				return emitErr
			}
		}
		return err
	}

	// Replay history accumulated so far
	history := make([]string, len(stream.history))
	copy(history, stream.history)
	stream.subscribers = append(stream.subscribers, sub)
	stream.mu.Unlock()
	sf.mu.Unlock()

	defer func() {
		stream.mu.Lock()
		for i, s := range stream.subscribers {
			if s == sub {
				stream.subscribers = append(stream.subscribers[:i], stream.subscribers[i+1:]...)
				break
			}
		}
		stream.mu.Unlock()
	}()

	for _, delta := range history {
		if emitErr := emit(delta); emitErr != nil {
			return emitErr
		}
	}

	// Stream new live deltas as they arrive
	for delta := range sub {
		if emitErr := emit(delta); emitErr != nil {
			return emitErr
		}
	}

	stream.mu.Lock()
	defer stream.mu.Unlock()
	return stream.err
}
