package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"yadro-go-course/internal/contextutil"
	"yadro-go-course/pkg/ratelimiter"
)

var dummyHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func TestRateLimitOnID(t *testing.T) {
	limiter := ratelimiter.NewRateLimiterPerUser[int64](1, time.Nanosecond)
	m := RateLimitOnID(limiter)
	h := m(dummyHandler)
	ctxFirst := contextutil.WithReqID(contextutil.WithUserID(context.Background(), 1), 1)
	ctxSecond := contextutil.WithReqID(contextutil.WithUserID(context.Background(), 2), 2)
	firstReq := httptest.NewRequest("GET", "/", nil).WithContext(ctxFirst)
	secondReq := httptest.NewRequest("GET", "/", nil).WithContext(ctxSecond)

	wFirst := httptest.NewRecorder()
	h.ServeHTTP(wFirst, firstReq)
	assert.Equal(t, http.StatusOK, wFirst.Result().StatusCode)
	wFirst = httptest.NewRecorder()
	h.ServeHTTP(wFirst, firstReq)
	assert.Equal(t, http.StatusTooManyRequests, wFirst.Result().StatusCode)

	wSecond := httptest.NewRecorder()
	h.ServeHTTP(wSecond, secondReq)
	assert.Equal(t, http.StatusOK, wSecond.Result().StatusCode)
	wSecond = httptest.NewRecorder()
	h.ServeHTTP(wSecond, secondReq)
	assert.Equal(t, http.StatusTooManyRequests, wSecond.Result().StatusCode)

	limiter.CleanUp()

	wFirst = httptest.NewRecorder()
	h.ServeHTTP(wFirst, firstReq)
	assert.Equal(t, http.StatusOK, wFirst.Result().StatusCode)

	wSecond = httptest.NewRecorder()
	h.ServeHTTP(wSecond, secondReq)
	assert.Equal(t, http.StatusOK, wSecond.Result().StatusCode)
}

func TestRateLimitOnIP(t *testing.T) {
	ctx := contextutil.WithReqID(context.Background(), 1)
	limiter := ratelimiter.NewRateLimiterPerUser[string](1, time.Nanosecond)
	m := RateLimitOnIP(limiter)
	h := m(dummyHandler)

	firstReq := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	firstReq.RemoteAddr = "first"
	secondReq := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	secondReq.RemoteAddr = "second"

	wFirst := httptest.NewRecorder()
	h.ServeHTTP(wFirst, firstReq)
	assert.Equal(t, http.StatusOK, wFirst.Result().StatusCode)
	wFirst = httptest.NewRecorder()
	h.ServeHTTP(wFirst, firstReq)
	assert.Equal(t, http.StatusTooManyRequests, wFirst.Result().StatusCode)

	wSecond := httptest.NewRecorder()
	h.ServeHTTP(wSecond, secondReq)
	assert.Equal(t, http.StatusOK, wSecond.Result().StatusCode)
	wSecond = httptest.NewRecorder()
	h.ServeHTTP(wSecond, secondReq)
	assert.Equal(t, http.StatusTooManyRequests, wSecond.Result().StatusCode)

	limiter.CleanUp()

	wFirst = httptest.NewRecorder()
	h.ServeHTTP(wFirst, firstReq)
	assert.Equal(t, http.StatusOK, wFirst.Result().StatusCode)

	wSecond = httptest.NewRecorder()
	h.ServeHTTP(wSecond, secondReq)
	assert.Equal(t, http.StatusOK, wSecond.Result().StatusCode)
}

func TestConcurrencyLimiter_Allowed(t *testing.T) {
	m := ConcurrencyLimiter(5)
	h := m(dummyHandler)
	ctx := contextutil.WithReqID(context.Background(), 1)
	r := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
}

func TestConcurrencyLimiter_Exceeded(t *testing.T) {
	m := ConcurrencyLimiter(0)
	h := m(dummyHandler)
	ctx := contextutil.WithReqID(context.Background(), 1)
	r := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusTooManyRequests, w.Result().StatusCode)
}

func TestConcurrencyLimiter_Sequential(t *testing.T) {
	m := ConcurrencyLimiter(2)
	h := m(dummyHandler)
	ctx := contextutil.WithReqID(context.Background(), 1)
	r := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	}
}
