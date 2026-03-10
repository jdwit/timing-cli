package api

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testServer(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client := NewClient("test-key", "Europe/Amsterdam")
	client.BaseURL = srv.URL
	return client
}

// --- APIError ---

func TestAPIError_WithMessage(t *testing.T) {
	err := &APIError{StatusCode: 422, Message: "Validation failed"}
	assert.Equal(t, "API error 422: Validation failed", err.Error())
}

func TestAPIError_WithoutMessage(t *testing.T) {
	err := &APIError{StatusCode: 500}
	assert.Equal(t, "API error 500", err.Error())
}

// --- NewClient ---

func TestNewClient(t *testing.T) {
	c := NewClient("key", "UTC")
	assert.Equal(t, "key", c.apiKey)
	assert.Equal(t, "UTC", c.timezone)
	assert.Equal(t, "https://web.timingapp.com/api/v1", c.BaseURL)
	assert.NotNil(t, c.httpClient)
}

// --- newRequest ---

func TestNewRequest_Headers(t *testing.T) {
	c := NewClient("my-api-key", "Europe/Amsterdam")
	req, err := c.newRequest("GET", "/test", nil)
	require.NoError(t, err)

	assert.Equal(t, "Bearer my-api-key", req.Header.Get("Authorization"))
	assert.Equal(t, "application/json", req.Header.Get("Accept"))
	assert.Equal(t, "Europe/Amsterdam", req.Header.Get("X-Time-Zone"))
	assert.Equal(t, "gzip", req.Header.Get("Accept-Encoding"))
}

func TestNewRequest_NoTimezone(t *testing.T) {
	c := NewClient("key", "")
	req, err := c.newRequest("GET", "/test", nil)
	require.NoError(t, err)
	assert.Empty(t, req.Header.Get("X-Time-Zone"))
}

func TestNewRequest_WithBody(t *testing.T) {
	c := NewClient("key", "UTC")
	body := bytes.NewReader([]byte(`{"title":"test"}`))
	req, err := c.newRequest("POST", "/test", body)
	require.NoError(t, err)
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
}

func TestNewRequest_WithoutBody(t *testing.T) {
	c := NewClient("key", "UTC")
	req, err := c.newRequest("GET", "/test", nil)
	require.NoError(t, err)
	assert.Empty(t, req.Header.Get("Content-Type"))
}

// --- Get ---

func TestGet_Success(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		w.Write([]byte(`{"data": [{"id": 1}]}`))
	})

	body, err := client.Get("/test", nil)
	require.NoError(t, err)

	var resp struct {
		Data []struct {
			ID int `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(body, &resp))
	require.Len(t, resp.Data, 1)
	assert.Equal(t, 1, resp.Data[0].ID)
}

func TestGet_WithParams(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "hello", r.URL.Query().Get("q"))
		w.Write([]byte(`{}`))
	})

	_, err := client.Get("/search", url.Values{"q": {"hello"}})
	require.NoError(t, err)
}

func TestGet_APIError_JSONMessage(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(422)
		w.Write([]byte(`{"message": "Validation failed"}`))
	})

	_, err := client.Get("/test", nil)
	require.Error(t, err)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 422, apiErr.StatusCode)
	assert.Equal(t, "Validation failed", apiErr.Message)
}

func TestGet_APIError_PlainBody(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`Internal Server Error`))
	})

	_, err := client.Get("/test", nil)
	require.Error(t, err)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 500, apiErr.StatusCode)
	assert.Equal(t, "Internal Server Error", apiErr.Message)
}

func TestGet_GzipResponse(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		gz.Write([]byte(`{"compressed": true}`))
		gz.Close()
	})

	body, err := client.Get("/test", nil)
	require.NoError(t, err)
	assert.JSONEq(t, `{"compressed": true}`, string(body))
}

// --- Rate limiting ---

func TestGet_RateLimitRetry(t *testing.T) {
	attempts := 0
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(429)
			return
		}
		w.Write([]byte(`{"ok": true}`))
	})

	body, err := client.Get("/test", nil)
	require.NoError(t, err)
	assert.JSONEq(t, `{"ok": true}`, string(body))
	assert.Equal(t, 2, attempts)
}

func TestGet_RateLimitExhausted(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(429)
	})

	_, err := client.Get("/test", nil)
	require.Error(t, err)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 429, apiErr.StatusCode)
}

// --- GetText ---

func TestGetText_Success(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "text/plain", r.Header.Get("Accept"))
		w.Write([]byte("hello world"))
	})

	text, err := client.GetText("/test", nil)
	require.NoError(t, err)
	assert.Equal(t, "hello world", text)
}

func TestGetText_WithParams(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "bar", r.URL.Query().Get("foo"))
		w.Write([]byte("ok"))
	})

	text, err := client.GetText("/test", url.Values{"foo": {"bar"}})
	require.NoError(t, err)
	assert.Equal(t, "ok", text)
}

func TestGetText_Error(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		w.Write([]byte("Forbidden"))
	})

	_, err := client.GetText("/test", nil)
	require.Error(t, err)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 403, apiErr.StatusCode)
	assert.Equal(t, "Forbidden", apiErr.Message)
}

// --- GetPaginated ---

func TestGetPaginated_MultiplePages(t *testing.T) {
	page := 0
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		page++
		json.NewEncoder(w).Encode(map[string]any{
			"data":  []map[string]int{{"id": page}},
			"meta":  map[string]int{"current_page": page, "last_page": 2, "total": 2},
			"links": map[string]any{},
		})
	})

	items, err := client.GetPaginated("/test", nil, 0)
	require.NoError(t, err)
	assert.Len(t, items, 2)
}

func TestGetPaginated_PageLimit(t *testing.T) {
	page := 0
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		page++
		json.NewEncoder(w).Encode(map[string]any{
			"data":  []map[string]int{{"id": page}},
			"meta":  map[string]int{"current_page": page, "last_page": 5, "total": 5},
			"links": map[string]any{},
		})
	})

	items, err := client.GetPaginated("/test", nil, 2)
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, 2, page)
}

func TestGetPaginated_SingleObject(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"data":  map[string]int{"id": 1},
			"meta":  map[string]int{"current_page": 1, "last_page": 1, "total": 1},
			"links": map[string]any{},
		})
	})

	items, err := client.GetPaginated("/test", nil, 0)
	require.NoError(t, err)
	assert.Len(t, items, 1)
}

func TestGetPaginated_Error(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"message": "server error"}`))
	})

	_, err := client.GetPaginated("/test", nil, 0)
	require.Error(t, err)
}

func TestGetPaginated_WithExistingParams(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "true", r.URL.Query().Get("include_project_data"))
		json.NewEncoder(w).Encode(map[string]any{
			"data":  []map[string]int{{"id": 1}},
			"meta":  map[string]int{"current_page": 1, "last_page": 1, "total": 1},
			"links": map[string]any{},
		})
	})

	params := url.Values{"include_project_data": {"true"}}
	items, err := client.GetPaginated("/test", params, 0)
	require.NoError(t, err)
	assert.Len(t, items, 1)
}

// --- Post ---

func TestPost_Success(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, _ := io.ReadAll(r.Body)
		var payload map[string]string
		require.NoError(t, json.Unmarshal(body, &payload))
		assert.Equal(t, "test", payload["title"])

		w.WriteHeader(201)
		w.Write([]byte(`{"data": {"self": "/projects/1", "title": "test"}}`))
	})

	resp, err := client.Post("/projects", bytes.NewReader([]byte(`{"title":"test"}`)))
	require.NoError(t, err)
	assert.Contains(t, string(resp), "/projects/1")
}

func TestPost_Error(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(422)
		w.Write([]byte(`{"message": "title required"}`))
	})

	_, err := client.Post("/projects", bytes.NewReader([]byte(`{}`)))
	require.Error(t, err)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 422, apiErr.StatusCode)
	assert.Equal(t, "title required", apiErr.Message)
}

func TestPost_NilBody(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.Write([]byte(`{"data": {}}`))
	})

	_, err := client.Post("/test", nil)
	require.NoError(t, err)
}

// --- Put ---

func TestPut_Success(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.Write([]byte(`{"data": {"self": "/projects/1", "title": "updated"}}`))
	})

	resp, err := client.Put("/projects/1", bytes.NewReader([]byte(`{"title":"updated"}`)))
	require.NoError(t, err)
	assert.Contains(t, string(resp), "updated")
}

func TestPut_Error(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"message": "not found"}`))
	})

	_, err := client.Put("/projects/999", bytes.NewReader([]byte(`{}`)))
	require.Error(t, err)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 404, apiErr.StatusCode)
}

func TestPut_NilBody(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data": {"self": "/time-entries/1"}}`))
	})

	resp, err := client.Put("/time-entries/stop", nil)
	require.NoError(t, err)
	assert.Contains(t, string(resp), "/time-entries/1")
}

// --- Patch ---

func TestPatch_Success(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		w.Write([]byte(`{"data": [{"self": "/time-entries/1"}]}`))
	})

	resp, err := client.Patch("/time-entries/batch-update", bytes.NewReader([]byte(`{}`)))
	require.NoError(t, err)
	assert.Contains(t, string(resp), "/time-entries/1")
}

func TestPatch_Error(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte(`{"message": "bad request"}`))
	})

	_, err := client.Patch("/test", bytes.NewReader([]byte(`{}`)))
	require.Error(t, err)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 400, apiErr.StatusCode)
	assert.Equal(t, "bad request", apiErr.Message)
}

// --- Delete ---

func TestDelete_Success(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(204)
	})

	err := client.Delete("/projects/1")
	assert.NoError(t, err)
}

func TestDelete_Error(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"message": "not found"}`))
	})

	err := client.Delete("/projects/999")
	require.Error(t, err)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 404, apiErr.StatusCode)
	assert.Equal(t, "not found", apiErr.Message)
}

func TestDelete_ErrorPlainBody(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`server error`))
	})

	err := client.Delete("/test")
	require.Error(t, err)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 500, apiErr.StatusCode)
}

// --- GetWithRedirect ---

func TestGetWithRedirect_Success(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data": {"self": "/time-entries/1", "is_running": true}}`))
	})

	body, err := client.GetWithRedirect("/time-entries/running", nil)
	require.NoError(t, err)
	require.NotNil(t, body)
	assert.Contains(t, string(body), "is_running")
}

func TestGetWithRedirect_404ReturnsNil(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})

	body, err := client.GetWithRedirect("/time-entries/running", nil)
	require.NoError(t, err)
	assert.Nil(t, body)
}

func TestGetWithRedirect_Error(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"message": "server error"}`))
	})

	_, err := client.GetWithRedirect("/test", nil)
	require.Error(t, err)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 500, apiErr.StatusCode)
}

func TestGetWithRedirect_WithParams(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "bar", r.URL.Query().Get("foo"))
		w.Write([]byte(`{}`))
	})

	body, err := client.GetWithRedirect("/test", url.Values{"foo": {"bar"}})
	require.NoError(t, err)
	assert.NotNil(t, body)
}

func TestGetWithRedirect_ErrorPlainBody(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		w.Write([]byte(`Forbidden`))
	})

	_, err := client.GetWithRedirect("/test", nil)
	require.Error(t, err)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "Forbidden", apiErr.Message)
}

// --- ExtractID ---

func TestExtractID(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"/projects/1", "1"},
		{"/time-entries/42", "42"},
		{"/teams/5", "5"},
		{"123", "123"},
		{"", ""},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, ExtractID(tt.input), "ExtractID(%q)", tt.input)
	}
}
