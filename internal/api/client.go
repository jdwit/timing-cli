package api

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	httpClient *http.Client
	apiKey     string
	timezone   string
	BaseURL    string
}

func NewClient(apiKey, timezone string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		apiKey:     apiKey,
		timezone:   timezone,
		BaseURL:    "https://web.timingapp.com/api/v1",
	}
}

type APIError struct {
	StatusCode int
	Message    string
	RetryAfter time.Duration
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("API error %d", e.StatusCode)
}

type PaginatedResponse struct {
	Data  json.RawMessage `json:"data"`
	Links struct {
		Next *string `json:"next"`
	} `json:"links"`
	Meta struct {
		CurrentPage int `json:"current_page"`
		LastPage    int `json:"last_page"`
		Total       int `json:"total"`
	} `json:"meta"`
}

func (c *Client) newRequest(method, path string, body io.Reader) (*http.Request, error) {
	u := c.BaseURL + path
	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")
	if c.timezone != "" {
		req.Header.Set("X-Time-Zone", c.timezone)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	maxRetries := 3
	for attempt := range maxRetries {
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request failed: %w", err)
		}

		if resp.StatusCode == 429 {
			resp.Body.Close()
			wait := time.Duration(30*(attempt+1)) * time.Second
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				if secs, err := strconv.Atoi(ra); err == nil {
					wait = time.Duration(secs) * time.Second
				}
			}
			if attempt < maxRetries-1 {
				time.Sleep(wait)
				// Clone the request for retry
				newReq := req.Clone(req.Context())
				req = newReq
				continue
			}
			return nil, &APIError{StatusCode: 429, Message: "rate limited, please try again later"}
		}

		return resp, nil
	}
	return nil, fmt.Errorf("max retries exceeded")
}

func (c *Client) readBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gr, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("decompressing response: %w", err)
		}
		defer gr.Close()
		reader = gr
	}

	return io.ReadAll(reader)
}

// Get performs a GET request and returns the response body.
func (c *Client) Get(path string, params url.Values) ([]byte, error) {
	if params != nil {
		path += "?" + params.Encode()
	}

	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}

	body, err := c.readBody(resp)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		msg := string(body)
		var errResp struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(body, &errResp) == nil && errResp.Message != "" {
			msg = errResp.Message
		}
		return nil, &APIError{StatusCode: resp.StatusCode, Message: msg}
	}

	return body, nil
}

// GetText performs a GET request expecting text/plain response.
func (c *Client) GetText(path string, params url.Values) (string, error) {
	if params != nil {
		path += "?" + params.Encode()
	}

	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "text/plain")

	resp, err := c.do(req)
	if err != nil {
		return "", err
	}

	body, err := c.readBody(resp)
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 400 {
		return "", &APIError{StatusCode: resp.StatusCode, Message: string(body)}
	}

	return string(body), nil
}

// GetPaginated performs paginated GET requests and collects all data.
func (c *Client) GetPaginated(path string, params url.Values, pageLimit int) ([]json.RawMessage, error) {
	if params == nil {
		params = url.Values{}
	}

	var allData []json.RawMessage
	page := 1

	for {
		params.Set("page", strconv.Itoa(page))
		body, err := c.Get(path, params)
		if err != nil {
			return nil, err
		}

		var paginated PaginatedResponse
		if err := json.Unmarshal(body, &paginated); err != nil {
			return nil, fmt.Errorf("parsing paginated response: %w", err)
		}

		// Each item in the data array
		var items []json.RawMessage
		if err := json.Unmarshal(paginated.Data, &items); err != nil {
			// Might be a single object, not an array
			allData = append(allData, paginated.Data)
			break
		}
		allData = append(allData, items...)

		if paginated.Meta.CurrentPage >= paginated.Meta.LastPage {
			break
		}
		if pageLimit > 0 && page >= pageLimit {
			break
		}
		page++
	}

	return allData, nil
}

// Post performs a POST request with a JSON body.
func (c *Client) Post(path string, body io.Reader) ([]byte, error) {
	req, err := c.newRequest("POST", path, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}

	respBody, err := c.readBody(resp)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		msg := string(respBody)
		var errResp struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Message != "" {
			msg = errResp.Message
		}
		return nil, &APIError{StatusCode: resp.StatusCode, Message: msg}
	}

	return respBody, nil
}

// Put performs a PUT request with a JSON body.
func (c *Client) Put(path string, body io.Reader) ([]byte, error) {
	req, err := c.newRequest("PUT", path, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}

	respBody, err := c.readBody(resp)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		msg := string(respBody)
		var errResp struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Message != "" {
			msg = errResp.Message
		}
		return nil, &APIError{StatusCode: resp.StatusCode, Message: msg}
	}

	return respBody, nil
}

// Patch performs a PATCH request with a JSON body.
func (c *Client) Patch(path string, body io.Reader) ([]byte, error) {
	req, err := c.newRequest("PATCH", path, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}

	respBody, err := c.readBody(resp)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		msg := string(respBody)
		var errResp struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Message != "" {
			msg = errResp.Message
		}
		return nil, &APIError{StatusCode: resp.StatusCode, Message: msg}
	}

	return respBody, nil
}

// Delete performs a DELETE request.
func (c *Client) Delete(path string) error {
	req, err := c.newRequest("DELETE", path, nil)
	if err != nil {
		return err
	}

	resp, err := c.do(req)
	if err != nil {
		return err
	}

	body, err := c.readBody(resp)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		msg := string(body)
		var errResp struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(body, &errResp) == nil && errResp.Message != "" {
			msg = errResp.Message
		}
		return &APIError{StatusCode: resp.StatusCode, Message: msg}
	}

	return nil
}

// GetWithRedirect performs a GET that follows redirects and returns the final response body.
func (c *Client) GetWithRedirect(path string, params url.Values) ([]byte, error) {
	if params != nil {
		path += "?" + params.Encode()
	}

	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	// Use a client that follows redirects (default behavior)
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}

	body, err := c.readBody(resp)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		return nil, nil
	}

	if resp.StatusCode >= 400 {
		msg := string(body)
		var errResp struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(body, &errResp) == nil && errResp.Message != "" {
			msg = errResp.Message
		}
		return nil, &APIError{StatusCode: resp.StatusCode, Message: msg}
	}

	return body, nil
}

// ExtractID extracts the numeric ID from a self reference like "/projects/1".
func ExtractID(self string) string {
	parts := strings.Split(self, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return self
}
