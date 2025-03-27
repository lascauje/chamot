package ollama

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestChat(t *testing.T) {
	mockResponses := []Response{
		{Response: "Hi,", Done: false},
		{Response: "What can I help with?", Done: true},
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req Request
		_ = json.NewDecoder(r.Body).Decode(&req)

		for _, res := range mockResponses {
			resBytes, _ := json.Marshal(res)
			_, _ = w.Write(resBytes)
			_, _ = w.Write([]byte("\n"))

			time.Sleep(10 * time.Millisecond)
		}
	}))
	defer mockServer.Close()

	ollamaClient := NewOllama(mockServer.URL, "test-model", 100)

	in := make(chan string, 1)
	out := make(chan Response, len(mockResponses))
	stop := make(chan bool, 1)

	go func() {
		_ = ollamaClient.Chat(in, out, stop)
	}()

	tests := []struct {
		input            string
		expectedResponse []Response
		stop             bool
	}{
		{
			input:            "Hello...",
			expectedResponse: mockResponses,
			stop:             false,
		},
		{
			input:            "Hello and bye",
			expectedResponse: []Response{mockResponses[0]},
			stop:             true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			in <- tt.input

			if tt.stop {
				stop <- true
			}

			receivedResponses := []Response{}
			for range mockResponses {
				receivedResponses = append(receivedResponses, <-out)
			}

			for i, res := range tt.expectedResponse {
				if receivedResponses[i] != res {
					t.Errorf("Response %d: expected %v, got %v", i, res, receivedResponses[i])
				}
			}
		})
	}
}
