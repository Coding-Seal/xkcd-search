package contextutil

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReqID(t *testing.T) {
	ctx := context.Background()

	assert.Panics(t, func() { ReqID(ctx) })

	ctx = WithReqID(ctx, 42)

	assert.NotPanics(t, func() { ReqID(ctx) })
	assert.Equal(t, 42, ReqID(ctx))
}

func TestReqID_Zero(t *testing.T) {
	ctx := WithReqID(context.Background(), 0)
	assert.NotPanics(t, func() { ReqID(ctx) })
	assert.Equal(t, 0, ReqID(ctx))
}

func TestReqID_Negative(t *testing.T) {
	ctx := WithReqID(context.Background(), -5)
	assert.Equal(t, -5, ReqID(ctx))
}

func TestReqID_Overwrite(t *testing.T) {
	ctx := WithReqID(context.Background(), 1)
	ctx = WithReqID(ctx, 2)
	assert.Equal(t, 2, ReqID(ctx))
}

func TestContext_AllThreeValues_Coexist(t *testing.T) {
	ctx := context.Background()
	ctx = WithReqID(ctx, 100)
	ctx = WithUserID(ctx, 200)
	ctx = WithIsAdmin(ctx, true)

	assert.Equal(t, 100, ReqID(ctx))
	assert.Equal(t, int64(200), UserID(ctx))
	assert.True(t, IsAdmin(ctx))
}

func TestContext_AllThreeValues_IndependentKeys(t *testing.T) {
	// Setting only reqID must not satisfy userID or isAdmin reads
	ctx := WithReqID(context.Background(), 1)
	assert.Panics(t, func() { UserID(ctx) })
	assert.Panics(t, func() { IsAdmin(ctx) })
}

func TestContext_ReqID_DoesNotAffectUserID(t *testing.T) {
	ctx := WithReqID(context.Background(), 999)
	ctx = WithUserID(ctx, 42)
	assert.Equal(t, 999, ReqID(ctx))
	assert.Equal(t, int64(42), UserID(ctx))
}
