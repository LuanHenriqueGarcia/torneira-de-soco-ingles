package internal

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HTTPClient struct {
	Client    *http.Client
	UserAgent string
	Headers   map[string]string
	Proxy     string
}

type HTTPResponse struct {
	StatusCode    int
	Body          string
	Headers       http.Header
	ContentLength int64
	URL           string
	Duration      time.Duration
}

func NewHTTPClient(timeout int, followRedirects bool, proxy string) *HTTPClient {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	}

	if proxy != "" {
		proxyURL, err := url.Parse(proxy)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(timeout) * time.Second,
	}

	if !followRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return &HTTPClient{
		Client:    client,
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		Headers:   make(map[string]string),
	}
}

func (h *HTTPClient) SetHeader(key, value string) {
	h.Headers[key] = value
}

func (h *HTTPClient) Get(targetURL string) (*HTTPResponse, error) {
	return h.DoRequest("GET", targetURL, "")
}

func (h *HTTPClient) Post(targetURL string, body string) (*HTTPResponse, error) {
	return h.DoRequest("POST", targetURL, body)
}

func (h *HTTPClient) DoRequest(method, targetURL, body string) (*HTTPResponse, error) {
	start := time.Now()

	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, targetURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}

	req.Header.Set("User-Agent", h.UserAgent)

	for key, value := range h.Headers {
		req.Header.Set(key, value)
	}

	if method == "POST" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := h.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro na requisição: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %w", err)
	}

	duration := time.Since(start)

	return &HTTPResponse{
		StatusCode:    resp.StatusCode,
		Body:          string(respBody),
		Headers:       resp.Header,
		ContentLength: resp.ContentLength,
		URL:           targetURL,
		Duration:      duration,
	}, nil
}

func (h *HTTPClient) PostForm(targetURL string, data map[string]string) (*HTTPResponse, error) {
	formData := url.Values{}
	for key, value := range data {
		formData.Set(key, value)
	}
	return h.Post(targetURL, formData.Encode())
}

func (h *HTTPClient) CheckConnection(targetURL string) bool {
	resp, err := h.Get(targetURL)
	return err == nil && resp.StatusCode > 0
}

func ExtractForms(html string) []string {
	var forms []string
	start := 0
	for {
		formStart := strings.Index(html[start:], "<form")
		if formStart == -1 {
			break
		}
		formEnd := strings.Index(html[start+formStart:], "</form>")
		if formEnd == -1 {
			break
		}
		forms = append(forms, html[start+formStart:start+formStart+formEnd+7])
		start = start + formStart + formEnd + 7
	}
	return forms
}
