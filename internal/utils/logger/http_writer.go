package logger

import (
	"bytes"
	"net/http"
	"time"
)

type httpWriter struct {
	url    string
	client *http.Client
}

func NewHTTPWriter(url string) *httpWriter {
	return &httpWriter{
		url: url,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (w *httpWriter) Write(p []byte) (n int, err error) {
	req, err := http.NewRequest("POST", w.url, bytes.NewBuffer(p))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return len(p), nil
}
