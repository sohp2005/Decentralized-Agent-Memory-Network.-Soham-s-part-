package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

// HTTPCacheClient implements CacheService via HTTP calls to cache-service.
type HTTPCacheClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewHTTPCacheClient(baseURL string) *HTTPCacheClient {
	return &HTTPCacheClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

// Get returns (value, true) on cache hit, ("", false) on miss or error.
func (c *HTTPCacheClient) Get(key string) (string, bool) {
	u := fmt.Sprintf("%s/get?key=%s", c.BaseURL, url.QueryEscape(key))

	resp, err := c.HTTPClient.Get(u)
	if err != nil {
		log.Printf("[CACHE-CLIENT] Get network error: %v", err)
		return "", false
	}
	defer resp.Body.Close()

	// Our cache-service returns 404 on miss.
	if resp.StatusCode == http.StatusNotFound {
		return "", false
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[CACHE-CLIENT] Get unexpected status %d: %s", resp.StatusCode, string(body))
		return "", false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[CACHE-CLIENT] Get read body error: %v", err)
		return "", false
	}

	// cache-service returns plain text value
	return string(body), true
}

func (c *HTTPCacheClient) Put(key, value string) error {
	u := fmt.Sprintf("%s/put", c.BaseURL)

	payload := map[string]string{
		"key":   key,
		"value": value,
	}
	b, _ := json.Marshal(payload)

	resp, err := c.HTTPClient.Post(u, "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("cache put failed: status=%d body=%s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *HTTPCacheClient) Invalidate(key string) error {
	u := fmt.Sprintf("%s/invalidate?key=%s", c.BaseURL, url.QueryEscape(key))

	req, err := http.NewRequest(http.MethodDelete, u, nil)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("cache invalidate failed: status=%d body=%s", resp.StatusCode, string(body))
	}
	return nil
}
