package ratelimiter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPerUser_Allow(t *testing.T) {
	limiter := NewRateLimiterPerUser[int](1, 3)
	assert.True(t, limiter.Allow(5))
	assert.False(t, limiter.Allow(5))
}

func TestPerUser_CleanUp(t *testing.T) {
	limiter := NewRateLimiterPerUser[int](1, 3)
	limiter.Allow(5)
	rl := limiter.clients[5].limiter
	limiter.clients[5] = &client{limiter: rl, lastSeen: time.Now().Add(time.Duration(-5) * time.Second)}
	limiter.CleanUp()
	assert.NotContains(t, limiter.clients, 5)
}

func TestRateLimiter_AllowFirst(t *testing.T) {
	lim := NewRateLimiterPerUser[string](5, time.Second)
	assert.True(t, lim.Allow("user1"), "first request should always be allowed")
}

func TestRateLimiter_DifferentUsers_Independent(t *testing.T) {
	lim := NewRateLimiterPerUser[string](1, time.Second)
	assert.True(t, lim.Allow("alpha"))
	assert.True(t, lim.Allow("beta"), "different user should have own bucket")
}

func TestRateLimiter_CleanupDoesNotPanic(t *testing.T) {
	lim := NewRateLimiterPerUser[int](10, time.Nanosecond)
	lim.Allow(1)
	lim.Allow(2)
	assert.NotPanics(t, func() { lim.CleanUp() })
}
