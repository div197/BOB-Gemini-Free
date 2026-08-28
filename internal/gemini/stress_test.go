package gemini

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type countingSessionRequester struct {
	calls atomic.Int64
}

func (c *countingSessionRequester) Do(req *http.Request) (*http.Response, error) {
	c.calls.Add(1)
	time.Sleep(20 * time.Millisecond) // Simulate realistic network latency
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(`"SNlM0e":"satavik-at-token","cfb2h":"satavik-bl-build"`)),
	}, nil
}

func Test500ConcurrentStudentsThunderingHerdProtection(t *testing.T) {
	cache := NewCookieCache("")
	mockHTTP := &countingSessionRequester{}

	const studentCount = 500
	var wg sync.WaitGroup
	wg.Add(studentCount)

	startGate := make(chan struct{})

	for i := 0; i < studentCount; i++ {
		go func(id int) {
			defer wg.Done()
			<-startGate // Align all 500 students to hit at the exact same microsecond

			at, bl, _, _ := cache.GetSessionInfo(context.Background(), mockHTTP, "")
			if at != "satavik-at-token" || bl != "satavik-bl-build" {
				t.Errorf("student %d got unexpected tokens: %s, %s", id, at, bl)
			}
		}(i)
	}

	close(startGate) // Release all 500 students!
	wg.Wait()

	calls := mockHTTP.calls.Load()
	if calls != 1 {
		t.Fatalf("Thundering-herd failed: 500 concurrent students caused %d upstream /app requests, want exactly 1", calls)
	}
}

func Test500ConcurrentStudentsStreamFlightMultiplexing(t *testing.T) {
	flight := NewStreamFlight()

	const totalStudents = 500
	const promptGroups = 5 // 5 different classroom coding assignments, 100 students each
	studentsPerGroup := totalStudents / promptGroups

	var upstreamCalls atomic.Int64
	var wg sync.WaitGroup
	wg.Add(totalStudents)

	startGate := make(chan struct{})

	for g := 0; g < promptGroups; g++ {
		promptName := fmt.Sprintf("Assignment #%d: Create 2D CyberGame", g+1)
		key := flight.Key(promptName, 1, 4, nil)

		for s := 0; s < studentsPerGroup; s++ {
			go func(groupID, studentID int, promptKey string) {
				defer wg.Done()
				<-startGate

				var collectedChunks []string
				err := flight.ExecuteStream(promptKey, func(emit func(string) error) error {
					upstreamCalls.Add(1)
					for c := 0; c < 10; c++ {
						time.Sleep(5 * time.Millisecond)
						if err := emit(fmt.Sprintf("chunk-%d", c)); err != nil {
							return err
						}
					}
					return nil
				}, func(delta string) error {
					collectedChunks = append(collectedChunks, delta)
					return nil
				})

				if err != nil {
					t.Errorf("group %d student %d error: %v", groupID, studentID, err)
				}
				if len(collectedChunks) != 10 {
					t.Errorf("group %d student %d got %d chunks, want 10", groupID, studentID, len(collectedChunks))
				}
			}(g, s, key)
		}
	}

	close(startGate)
	wg.Wait()

	calls := upstreamCalls.Load()
	if calls != promptGroups {
		t.Fatalf("Multiplexing failed: 500 students across 5 prompt groups caused %d upstream calls, want exactly %d", calls, promptGroups)
	}
}
