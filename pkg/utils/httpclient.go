package utils

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

func NewHTTPClient(timeout, dialTimeout, headerTimeout time.Duration, maxConns int) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DialContext:           (&net.Dialer{Timeout: dialTimeout}).DialContext,
			ResponseHeaderTimeout: headerTimeout,
			MaxIdleConns:          maxConns,
			MaxIdleConnsPerHost:   maxConns,
			IdleConnTimeout:       30 * time.Second,
			DisableKeepAlives:     false,
		},
	}
}

func HTTPGet(ctx context.Context, client *http.Client, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		Debugf("http: GET %s failed: %v", url, err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		Debugf("http: GET %s returned %d", url, resp.StatusCode)
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	Debugf("http: GET %s status=200 content-length=%s", url, resp.Header.Get("Content-Length"))
	return resp.Body, nil
}
