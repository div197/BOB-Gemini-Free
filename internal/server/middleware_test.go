package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggingResponseWriterKeepsFirstCommittedStatus(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := newLoggingResponseWriter(recorder)

	writer.WriteHeader(http.StatusBadRequest)
	writer.WriteHeader(http.StatusInternalServerError)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("recorder status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if writer.statusCode != http.StatusBadRequest {
		t.Fatalf("logged status = %d, want first committed status %d", writer.statusCode, http.StatusBadRequest)
	}
}

func TestLoggingResponseWriterRecordsImplicitAndFlushedStatus(t *testing.T) {
	t.Run("write", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		writer := newLoggingResponseWriter(recorder)

		if _, err := writer.Write([]byte("ok")); err != nil {
			t.Fatalf("Write() error = %v", err)
		}

		if writer.statusCode != http.StatusOK {
			t.Fatalf("logged status after Write = %d, want %d", writer.statusCode, http.StatusOK)
		}
	})

	t.Run("flush", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		writer := newLoggingResponseWriter(recorder)

		writer.Flush()
		writer.WriteHeader(http.StatusServiceUnavailable)

		if writer.statusCode != http.StatusOK {
			t.Fatalf("logged status after Flush = %d, want %d", writer.statusCode, http.StatusOK)
		}
	})
}
