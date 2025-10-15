package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rishabh-sonic/orbit/internal/middleware"
)

type apiResp struct {
	Data    any    `json:"data"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

func decode(t *testing.T, rr *httptest.ResponseRecorder) apiResp {
	t.Helper()
	var r apiResp
	if err := json.NewDecoder(rr.Body).Decode(&r); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return r
}

func TestOk(t *testing.T) {
	rr := httptest.NewRecorder()
	middleware.Ok(rr, map[string]string{"key": "value"})
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type: got %q, want application/json", ct)
	}
}

func TestCreated(t *testing.T) {
	rr := httptest.NewRecorder()
	middleware.Created(rr, map[string]string{"id": "123"})
	if rr.Code != http.StatusCreated {
		t.Errorf("status: got %d, want 201", rr.Code)
	}
}

func TestNoContent(t *testing.T) {
	rr := httptest.NewRecorder()
	middleware.NoContent(rr)
	if rr.Code != http.StatusNoContent {
		t.Errorf("status: got %d, want 204", rr.Code)
	}
	if rr.Body.Len() != 0 {
		t.Error("expected empty body for NoContent")
	}
}

func TestBadRequest(t *testing.T) {
	rr := httptest.NewRecorder()
	middleware.BadRequest(rr, "invalid input")
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", rr.Code)
	}
	r := decode(t, rr)
	if r.Error != "invalid input" {
		t.Errorf("error: got %q, want 'invalid input'", r.Error)
	}
}

func TestUnauthorized(t *testing.T) {
	rr := httptest.NewRecorder()
	middleware.Unauthorized(rr, "not authenticated")
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", rr.Code)
	}
}

func TestForbidden(t *testing.T) {
	rr := httptest.NewRecorder()
	middleware.Forbidden(rr, "no access")
	if rr.Code != http.StatusForbidden {
		t.Errorf("status: got %d, want 403", rr.Code)
	}
	r := decode(t, rr)
	if r.Error != "no access" {
		t.Errorf("error: got %q, want 'no access'", r.Error)
	}
}

func TestNotFound(t *testing.T) {
	rr := httptest.NewRecorder()
	middleware.NotFound(rr, "item missing")
	if rr.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want 404", rr.Code)
	}
}

func TestConflict(t *testing.T) {
	rr := httptest.NewRecorder()
	middleware.Conflict(rr, "already exists")
	if rr.Code != http.StatusConflict {
		t.Errorf("status: got %d, want 409", rr.Code)
	}
	r := decode(t, rr)
	if r.Error != "already exists" {
		t.Errorf("error: got %q, want 'already exists'", r.Error)
	}
}

func TestInternalError(t *testing.T) {
	rr := httptest.NewRecorder()
	middleware.InternalError(rr, nil)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want 500", rr.Code)
	}
	r := decode(t, rr)
	if r.Error == "" {
		t.Error("expected non-empty error message for InternalError")
	}
}
