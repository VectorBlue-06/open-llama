package chat

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// CompletionRequest is the request body for the /completion endpoint.
type CompletionRequest struct {
	Prompt        string   `json:"prompt"`
	Stream        bool     `json:"stream"`
	Temperature   float64  `json:"temperature"`
	TopP          float64  `json:"top_p"`
	TopK          int      `json:"top_k"`
	RepeatPenalty float64  `json:"repeat_penalty"`
	NPredict      int      `json:"n_predict"`
	Stop          []string `json:"stop"`
}

// streamCompletion sends a streaming completion request and returns a channel of tokens.
func streamCompletion(ctx context.Context, client *http.Client, serverPort int, req CompletionRequest) (<-chan StreamToken, error) {
	ch := make(chan StreamToken, 64)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("http://127.0.0.1:%d/completion", serverPort),
		strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	go func() {
		defer close(ch)

		resp, err := client.Do(httpReq)
		if err != nil {
			ch <- StreamToken{Content: "[Error: " + err.Error() + "]", Stop: true}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			ch <- StreamToken{Content: fmt.Sprintf("[Error: server returned %d]", resp.StatusCode), Stop: true}
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")

			var token StreamToken
			if err := json.Unmarshal([]byte(data), &token); err != nil {
				continue
			}

			select {
			case ch <- token:
			case <-ctx.Done():
				return
			}

			if token.Stop {
				return
			}
		}
	}()

	return ch, nil
}
