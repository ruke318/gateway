package proxy

import (
	"bytes"
	"io"
	"net/http"
)

type Forwarder struct {
	backendURL string
	client     *http.Client
}

func NewForwarder(backendURL string) *Forwarder {
	return &Forwarder{
		backendURL: backendURL,
		client:     &http.Client{},
	}
}

func (f *Forwarder) Forward(req *http.Request, body []byte) (*http.Response, []byte, error) {
	return f.ForwardWithOptions(req.Method, f.backendURL, req.URL.Path, body, req.Header)
}

func (f *Forwarder) ForwardWithOptions(method, backendURL, path string, body []byte, headers http.Header) (*http.Response, []byte, error) {
	url := backendURL + path
	proxyReq, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}

	for k, v := range headers {
		proxyReq.Header[k] = v
	}

	resp, err := f.client.Do(proxyReq)
	if err != nil {
		return nil, nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, nil, err
	}

	return resp, respBody, nil
}
