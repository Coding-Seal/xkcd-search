package http_util

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"yadro-go-course/internal/contextutil"
)

func TestWrapHandler(t *testing.T) {
	ctx := contextutil.WithReqID(context.Background(), 1)
	h := WrapHandler(func(w http.ResponseWriter, r *http.Request) error {
		return ErrForbidden
	})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)

	h = WrapHandler(func(w http.ResponseWriter, r *http.Request) error {
		return ErrBadRequest
	})
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)

	h = WrapHandler(func(w http.ResponseWriter, r *http.Request) error {
		return ErrNotFound
	})
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)

	h = WrapHandler(func(w http.ResponseWriter, r *http.Request) error {
		return nil
	})
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
}

func TestWrapHandler_NotFound(t *testing.T) {
	ctx := contextutil.WithReqID(context.Background(), 1)
	h := WrapHandler(func(w http.ResponseWriter, r *http.Request) error {
		return ErrNotFound
	})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestWrapHandler_BadRequest(t *testing.T) {
	ctx := contextutil.WithReqID(context.Background(), 1)
	h := WrapHandler(func(w http.ResponseWriter, r *http.Request) error {
		return ErrBadRequest
	})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWrapHandler_NoError(t *testing.T) {
	ctx := contextutil.WithReqID(context.Background(), 1)
	h := WrapHandler(func(w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(http.StatusCreated)
		return nil
	})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestWriteJson(t *testing.T) {
	w := httptest.NewRecorder()
	err := WriteJson(w, map[string]int{"answer": 42})
	assert.NoError(t, err)
	assert.Contains(t, w.Body.String(), "42")
}

func TestWriteJson_Struct(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
		Val  int    `json:"val"`
	}
	w := httptest.NewRecorder()
	err := WriteJson(w, payload{Name: "test", Val: 7})
	assert.NoError(t, err)
	assert.Contains(t, w.Body.String(), "test")
	assert.Contains(t, w.Body.String(), "7")
}

func TestWriteJson_Unencodable(t *testing.T) {
	w := httptest.NewRecorder()
	// Channels cannot be JSON-encoded, so this triggers the error path
	err := WriteJson(w, make(chan int))
	assert.Error(t, err)
}
