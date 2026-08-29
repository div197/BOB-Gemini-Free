package gemini

import (
	"context"
	"errors"
	"fmt"
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

func TestStreamFlightScopeSeparatesRequestDomains(t *testing.T) {
	flight := NewStreamFlight()
	anonymous := flight.KeyWithScope("anonymous", "same prompt", 1, 4, nil)
	accountA := flight.KeyWithScope("account-a", "same prompt", 1, 4, nil)
	accountB := flight.KeyWithScope("account-b", "same prompt", 1, 4, nil)
	if anonymous == accountA || accountA == accountB {
		t.Fatal("different flight scopes unexpectedly shared a key")
	}
	if anonymous != flight.KeyWithScope("anonymous", "same prompt", 1, 4, nil) {
		t.Fatal("same flight scope did not produce a stable key")
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

func TestStreamFlightReportsSlowSubscriberInsteadOfDropping(t *testing.T) {
	flight := NewStreamFlight()
	key := flight.Key("slow-subscriber", 1, 4, nil)
	upstreamStarted := make(chan struct{})
	startEmitting := make(chan struct{})
	followerRelease := make(chan struct{})
	leaderDone := make(chan error, 1)
	followerDone := make(chan error, 1)

	go func() {
		leaderDone <- flight.ExecuteStreamContext(context.Background(), key, func(emit func(string) error) error {
			close(upstreamStarted)
			<-startEmitting
			for i := 0; i < maxStreamSubscriberBuffer+2; i++ {
				if err := emit(fmt.Sprintf("chunk-%03d", i)); err != nil {
					return err
				}
			}
			return nil
		}, func(string) error { return nil })
	}()

	<-upstreamStarted
	go func() {
		followerDone <- flight.ExecuteStreamContext(context.Background(), key, func(emit func(string) error) error {
			return nil
		}, func(delta string) error {
			if delta == "chunk-000" {
				// The callback intentionally stops consuming while the bounded
				// subscriber queue fills, exercising the explicit overflow path.
				<-followerRelease
			}
			return nil
		})
	}()

	waitForFlightSubscriber(t, flight, key)
	close(startEmitting)

	select {
	case err := <-leaderDone:
		if err != nil {
			t.Fatalf("leader returned unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("leader did not finish after subscriber overflow")
	}

	close(followerRelease)
	select {
	case err := <-followerDone:
		if !errors.Is(err, ErrStreamSubscriberTooSlow) {
			t.Fatalf("follower error = %v, want ErrStreamSubscriberTooSlow", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("slow follower did not receive an explicit completion error")
	}
}

func TestStreamFlightFollowerContextCancellation(t *testing.T) {
	flight := NewStreamFlight()
	key := flight.Key("cancel-follower", 1, 4, nil)
	upstreamStarted := make(chan struct{})
	releaseUpstream := make(chan struct{})
	leaderDone := make(chan error, 1)

	go func() {
		leaderDone <- flight.ExecuteStreamContext(context.Background(), key, func(emit func(string) error) error {
			close(upstreamStarted)
			<-releaseUpstream
			return emit("final")
		}, func(string) error { return nil })
	}()
	<-upstreamStarted

	ctx, cancel := context.WithCancel(context.Background())
	followerDone := make(chan error, 1)
	go func() {
		followerDone <- flight.ExecuteStreamContext(ctx, key, func(emit func(string) error) error {
			return emit("unexpected-history")
		}, func(string) error { return nil })
	}()
	waitForFlightSubscriber(t, flight, key)
	cancel()

	select {
	case err := <-followerDone:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("follower error = %v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("follower did not stop after context cancellation")
	}

	close(releaseUpstream)
	select {
	case err := <-leaderDone:
		if err != nil {
			t.Fatalf("leader returned unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("leader did not finish after follower cancellation")
	}
}

func TestStreamFlightLeaderContextCancellationPropagatesToStream(t *testing.T) {
	flight := NewStreamFlight()
	key := flight.Key("cancel-leader", 1, 4, nil)
	upstreamStarted := make(chan struct{})
	releaseUpstream := make(chan struct{})
	leaderCtx, cancelLeader := context.WithCancel(context.Background())
	defer cancelLeader()
	leaderDone := make(chan error, 1)

	go func() {
		leaderDone <- flight.ExecuteStreamContext(leaderCtx, key, func(emit func(string) error) error {
			close(upstreamStarted)
			<-releaseUpstream
			return emit("late")
		}, func(string) error { return nil })
	}()
	<-upstreamStarted

	followerDone := make(chan error, 1)
	go func() {
		followerDone <- flight.ExecuteStreamContext(context.Background(), key, func(func(string) error) error {
			return nil
		}, func(string) error { return nil })
	}()
	waitForFlightSubscriber(t, flight, key)

	cancelLeader()
	close(releaseUpstream)

	select {
	case err := <-leaderDone:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("leader error = %v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("leader did not stop after context cancellation")
	}
	select {
	case err := <-followerDone:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("follower error = %v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("follower did not receive the cancelled stream result")
	}
}

func TestStreamFlightRejectsNilCallbacks(t *testing.T) {
	flight := NewStreamFlight()
	if _, err := flight.ExecuteContext(context.Background(), "key", nil); !errors.Is(err, ErrStreamRunnerNil) {
		t.Fatalf("ExecuteContext(nil) error = %v, want %v", err, ErrStreamRunnerNil)
	}
	if err := flight.ExecuteStreamContext(context.Background(), "key", nil, func(string) error { return nil }); !errors.Is(err, ErrStreamRunnerNil) {
		t.Fatalf("ExecuteStreamContext(nil runner) error = %v, want %v", err, ErrStreamRunnerNil)
	}
	if err := flight.ExecuteStreamContext(context.Background(), "key", func(func(string) error) error { return nil }, nil); !errors.Is(err, ErrStreamEmitterNil) {
		t.Fatalf("ExecuteStreamContext(nil emitter) error = %v, want %v", err, ErrStreamEmitterNil)
	}
}

func TestStreamFlightRejectsNewSubscriberAfterHistoryLimit(t *testing.T) {
	flight := NewStreamFlight()
	key := flight.Key("history-limit", 1, 4, nil)
	historyLimitReached := make(chan struct{})
	releaseUpstream := make(chan struct{})
	leaderDone := make(chan error, 1)

	go func() {
		leaderDone <- flight.ExecuteStreamContext(context.Background(), key, func(emit func(string) error) error {
			for i := 0; i < maxStreamHistoryChunks+1; i++ {
				if err := emit("x"); err != nil {
					return err
				}
			}
			close(historyLimitReached)
			<-releaseUpstream
			return nil
		}, func(string) error { return nil })
	}()

	<-historyLimitReached
	err := flight.ExecuteStreamContext(context.Background(), key, func(func(string) error) error {
		return nil
	}, func(string) error { return nil })
	if !errors.Is(err, ErrStreamHistoryLimit) {
		t.Fatalf("new subscriber error = %v, want ErrStreamHistoryLimit", err)
	}

	close(releaseUpstream)
	select {
	case err := <-leaderDone:
		if err != nil {
			t.Fatalf("leader returned unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("leader did not finish after history-limit check")
	}
}

func waitForFlightSubscriber(t *testing.T, flight *StreamFlight, key string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		flight.mu.Lock()
		stream := flight.streams[key]
		flight.mu.Unlock()
		if stream != nil {
			stream.mu.Lock()
			subscribers := len(stream.subscribers)
			stream.mu.Unlock()
			if subscribers > 0 {
				return
			}
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("timed out waiting for StreamFlight follower subscription")
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
