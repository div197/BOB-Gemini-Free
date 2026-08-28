package gemini

import (
	"io"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestStreamFlightMultiplexesConcurrentSubscribers(t *testing.T) {
	flight := NewStreamFlight()
	key := flight.Key("create cybersnake 2d game", 1, 4, nil)

	var wg sync.WaitGroup
	subscriberCount := 5
	results := make([][]string, subscriberCount)

	// Stream generator that emits chunks with slight delay to allow subscribers to join
	runUpstream := func(emit func(string) error) error {
		chunks := []string{"<!DOCTYPE html>", "<html>", "<head>", "<title>CyberSnake</title>", "<body>", "<canvas></canvas>", "</body>", "</html>"}
		for _, ch := range chunks {
			time.Sleep(10 * time.Millisecond)
			if err := emit(ch); err != nil {
				return err
			}
		}
		return nil
	}

	for i := 0; i < subscriberCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			var collected []string
			err := flight.ExecuteStream(key, runUpstream, func(delta string) error {
				collected = append(collected, delta)
				return nil
			})
			if err != nil {
				t.Errorf("subscriber %d error: %v", idx, err)
			}
			results[idx] = collected
		}(i)
	}

	wg.Wait()

	want := []string{"<!DOCTYPE html>", "<html>", "<head>", "<title>CyberSnake</title>", "<body>", "<canvas></canvas>", "</body>", "</html>"}
	for i, got := range results {
		if !reflect.DeepEqual(got, want) {
			t.Errorf("subscriber %d got %#v, want %#v", i, got, want)
		}
	}
}

func TestStreamFlightHandlesEarlySubscriberDisconnect(t *testing.T) {
	flight := NewStreamFlight()
	key := flight.Key("game-early-cancel", 1, 4, nil)

	var leaderEmitted []string
	var followerEmitted []string
	var wg sync.WaitGroup

	runUpstream := func(emit func(string) error) error {
		for i := 0; i < 6; i++ {
			time.Sleep(10 * time.Millisecond)
			if err := emit(string(rune('A' + i))); err != nil {
				return err
			}
		}
		return nil
	}

	wg.Add(2)

	// Leader
	go func() {
		defer wg.Done()
		_ = flight.ExecuteStream(key, runUpstream, func(delta string) error {
			leaderEmitted = append(leaderEmitted, delta)
			return nil
		})
	}()

	// Follower that simulates disconnect after 2 chunks
	go func() {
		defer wg.Done()
		time.Sleep(5 * time.Millisecond)
		_ = flight.ExecuteStream(key, runUpstream, func(delta string) error {
			followerEmitted = append(followerEmitted, delta)
			if len(followerEmitted) >= 2 {
				return io.EOF // Simulate client disconnect
			}
			return nil
		})
	}()

	wg.Wait()

	if len(leaderEmitted) != 6 {
		t.Fatalf("leader did not complete all 6 chunks, got %d chunks: %#v", len(leaderEmitted), leaderEmitted)
	}
	if len(followerEmitted) != 2 {
		t.Fatalf("follower was supposed to abort at 2 chunks, got %d chunks: %#v", len(followerEmitted), followerEmitted)
	}
}

func TestStreamFlight1000GoroutineChaosFuzz(t *testing.T) {
	flight := NewStreamFlight()
	var wg sync.WaitGroup
	goroutineCount := 1000
	wg.Add(goroutineCount)

	keys := []string{
		flight.Key("prompt A", 1, 0, nil),
		flight.Key("prompt B", 2, 4, nil),
		flight.Key("prompt C", 3, 2, nil),
		flight.Key("prompt D", 1, 4, nil),
	}

	runUpstreamStream := func(emit func(string) error) error {
		for i := 0; i < 5; i++ {
			time.Sleep(2 * time.Millisecond)
			if err := emit("token"); err != nil {
				return err
			}
		}
		return nil
	}

	runUpstreamNonStream := func() (string, error) {
		time.Sleep(5 * time.Millisecond)
		return "complete-response", nil
	}

	for i := 0; i < goroutineCount; i++ {
		go func(idx int) {
			defer wg.Done()
			k := keys[idx%len(keys)]
			if idx%2 == 0 {
				// Stream subscriber
				_ = flight.ExecuteStream(k, runUpstreamStream, func(delta string) error {
					if idx%5 == 0 {
						return io.EOF // Random early disconnect
					}
					return nil
				})
			} else {
				// Non-stream subscriber
				_, _ = flight.Execute(k, runUpstreamNonStream)
			}
		}(i)
	}

	wg.Wait()
}
