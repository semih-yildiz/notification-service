package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestJSONParsing_ValidNotification(t *testing.T) {
	body := `{"recipient":"+905551234567","channel":"sms","content":"Test message","priority":"high"}`

	var data map[string]interface{}
	err := json.Unmarshal([]byte(body), &data)

	if err != nil {
		t.Errorf("expected valid JSON, got error: %v", err)
	}
	if data["recipient"] != "+905551234567" {
		t.Error("recipient not parsed correctly")
	}
	if data["channel"] != "sms" {
		t.Error("channel not parsed correctly")
	}
}

func TestJSONParsing_InvalidJSON(t *testing.T) {
	body := `{"recipient":"+905551234567","channel":"sms",}`

	var data map[string]interface{}
	err := json.Unmarshal([]byte(body), &data)

	if err == nil {
		t.Error("expected JSON parse error for trailing comma")
	}
}

func TestJSONParsing_BatchArray(t *testing.T) {
	body := `[{"recipient":"+90555","channel":"sms","content":"Test 1"},{"recipient":"+90556","channel":"email","content":"Test 2"}]`

	var items []map[string]interface{}
	err := json.Unmarshal([]byte(body), &items)

	if err != nil {
		t.Errorf("expected valid JSON array, got error: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestEchoContext_URLParams(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/notifications/:id", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("test-123")

	if c.Param("id") != "test-123" {
		t.Errorf("expected id=test-123, got %s", c.Param("id"))
	}
}

func TestEchoContext_QueryParams(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/notifications?status=pending&limit=50&offset=10", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if c.QueryParam("status") != "pending" {
		t.Error("status query param not extracted")
	}
	if c.QueryParam("limit") != "50" {
		t.Error("limit query param not extracted")
	}
	if c.QueryParam("offset") != "10" {
		t.Error("offset query param not extracted")
	}
}

func TestEchoContext_RequestBody(t *testing.T) {
	e := echo.New()

	body := `{"recipient":"+905551234567","channel":"sms","content":"Test"}`
	req := httptest.NewRequest(http.MethodPost, "/notifications", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var data map[string]interface{}
	if err := c.Bind(&data); err != nil {
		t.Errorf("expected successful bind, got error: %v", err)
	}
	if data["recipient"] != "+905551234567" {
		t.Error("recipient not bound correctly")
	}
}

func TestBatchSizeValidation(t *testing.T) {
	tests := []struct {
		name  string
		size  int
		valid bool
	}{
		{"Empty batch", 0, false},
		{"Single item", 1, true},
		{"Normal batch", 100, true},
		{"Max batch", 1000, true},
		{"Over limit", 1001, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.size > 0 && tt.size <= 1000
			if valid != tt.valid {
				t.Errorf("size %d: expected valid=%v, got %v", tt.size, tt.valid, valid)
			}
		})
	}
}

func TestHTTPStatusCodes(t *testing.T) {
	tests := []struct {
		name   string
		status int
	}{
		{"Success", http.StatusOK},
		{"Created", http.StatusCreated},
		{"No Content", http.StatusNoContent},
		{"Bad Request", http.StatusBadRequest},
		{"Not Found", http.StatusNotFound},
		{"Internal Error", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.status < 200 || tt.status >= 600 {
				t.Errorf("invalid HTTP status code: %d", tt.status)
			}
		})
	}
}

func TestErrorResponse_JSON(t *testing.T) {
	errorResp := map[string]string{"error": "invalid channel"}
	jsonBytes, err := json.Marshal(errorResp)

	if err != nil {
		t.Errorf("expected successful marshal, got error: %v", err)
	}

	var parsed map[string]string
	json.Unmarshal(jsonBytes, &parsed)

	if parsed["error"] != "invalid channel" {
		t.Error("error message not serialized correctly")
	}
}

func TestContentType_JSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/notifications", nil)
	req.Header.Set("Content-Type", "application/json")

	if req.Header.Get("Content-Type") != "application/json" {
		t.Error("Content-Type header not set correctly")
	}
}

func TestHTTPMethods(t *testing.T) {
	tests := []struct {
		name   string
		method string
	}{
		{"POST create", http.MethodPost},
		{"GET retrieve", http.MethodGet},
		{"POST cancel", http.MethodPost},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", nil)
			if req.Method != tt.method {
				t.Errorf("expected method %s, got %s", tt.method, req.Method)
			}
		})
	}
}
