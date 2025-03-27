package ollama

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type API interface {
	Chat(in chan string, out chan Response, stop chan bool) error
}

type Ollama struct {
	url    string
	model  string
	numCtx int
	client *http.Client
}

type Options struct {
	NumCtx int `json:"num_ctx"` //nolint:tagliatelle // Well it's Ollama
}

type Request struct {
	Model   string  `json:"model"`
	Prompt  string  `json:"prompt"`
	Stream  bool    `json:"stream"`
	Options Options `json:"options"`
}

type Response struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func NewOllama(url string, model string, numCtx int) *Ollama {
	return &Ollama{
		url:    url,
		model:  model,
		numCtx: numCtx,
		client: &http.Client{
			Transport:     nil,
			CheckRedirect: nil,
			Jar:           nil,
			Timeout:       0,
		},
	}
}

func (o *Ollama) Chat(in chan string, out chan Response, stop chan bool) error {
	for {
		message := <-in

		if len(message) > o.numCtx {
			half := o.numCtx / 2
			message = message[:half] + "\n...\n" + message[len(message)-half:]
		}

		body, err := json.Marshal(Request{
			Model:   o.model,
			Prompt:  message,
			Stream:  true,
			Options: Options{NumCtx: o.numCtx},
		})
		if err != nil {
			return fmt.Errorf("error marshaling: %w", err)
		}

		resp, err := o.client.Post(o.url, "application/json", bytes.NewBuffer(body))
		if err != nil {
			return fmt.Errorf("error sending: %w", err)
		}

		defer func() {
			_ = resp.Body.Close()
		}()

		scanner := bufio.NewScanner(resp.Body)
		done := false

		for scanner.Scan() {
			var ollamaResp Response
			if err := json.Unmarshal(scanner.Bytes(), &ollamaResp); err == nil {
				out <- ollamaResp
			}
			select {
			case <-stop:
				done = true
			default:
			}

			if done {
				_ = resp.Body.Close()
				out <- Response{Response: "", Done: true}

				break
			}
		}
	}
}
