package contextutil

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAdmin(t *testing.T) {
	ctx := context.Background()

	assert.Panics(t, func() { IsAdmin(ctx) })

	ctx = WithIsAdmin(ctx, true)

	assert.NotPanics(t, func() { IsAdmin(ctx) })
	assert.True(t, IsAdmin(ctx))
}

func TestIsAdmin_False(t *testing.T) {
	ctx := WithIsAdmin(context.Background(), false)
	assert.NotPanics(t, func() { IsAdmin(ctx) })
	assert.False(t, IsAdmin(ctx))
}

func TestIsAdmin_Overwrite(t *testing.T) {
	ctx := WithIsAdmin(context.Background(), true)
	assert.True(t, IsAdmin(ctx))
	ctx = WithIsAdmin(ctx, false)
	assert.False(t, IsAdmin(ctx))
}

func TestUserID(t *testing.T) {
	ctx := context.Background()

	assert.Panics(t, func() { UserID(ctx) })

	ctx = WithUserID(ctx, 42)

	assert.NotPanics(t, func() { UserID(ctx) })
	assert.Equal(t, int64(42), UserID(ctx))
}

func TestUserID_Zero(t *testing.T) {
	ctx := WithUserID(context.Background(), 0)
	assert.NotPanics(t, func() { UserID(ctx) })
	assert.Equal(t, int64(0), UserID(ctx))
}

func TestUserID_Negative(t *testing.T) {
	ctx := WithUserID(context.Background(), -1)
	assert.Equal(t, int64(-1), UserID(ctx))
}

func TestUserID_Overwrite(t *testing.T) {
	ctx := WithUserID(context.Background(), 1)
	ctx = WithUserID(ctx, 99)
	assert.Equal(t, int64(99), UserID(ctx))
}

func TestContext_UserIDAndIsAdmin_Coexist(t *testing.T) {
	ctx := context.Background()
	ctx = WithUserID(ctx, 7)
	ctx = WithIsAdmin(ctx, true)

	assert.Equal(t, int64(7), UserID(ctx))
	assert.True(t, IsAdmin(ctx))
}

func TestContext_UserIDAndIsAdmin_IndependentKeys(t *testing.T) {
	// Verify the two keys don't interfere with each other
	ctx := WithUserID(context.Background(), 100)
	assert.Panics(t, func() { IsAdmin(ctx) }, "isAdmin not set yet, must panic")

	ctx2 := WithIsAdmin(context.Background(), false)
	assert.Panics(t, func() { UserID(ctx2) }, "userID not set yet, must panic")
}
