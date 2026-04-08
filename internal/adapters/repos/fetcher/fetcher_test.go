package fetcher

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	testfetcher "yadro-go-course/test/fetcher"
)

func newMock(t *testing.T, lastID int) (*httptest.Server, *Fetcher) {
	t.Helper()
	srv := testfetcher.NewMockXKCD(lastID)
	t.Cleanup(srv.Close)
	return srv, NewFetcher(srv.URL, 10)
}

func TestFetcher_LastComicID(t *testing.T) {
	_, f := newMock(t, 20)
	_, err := f.LastComicID(context.Background())
	assert.NoError(t, err)
}

func TestFetcher_LastComicID_ReturnsCorrectValue(t *testing.T) {
	_, f := newMock(t, 42)
	id, err := f.LastComicID(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 42, id)
}

func TestFetcher_LastComicID_One(t *testing.T) {
	_, f := newMock(t, 1)
	id, err := f.LastComicID(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, id)
}

func TestFetcher_LastComicID_UnreachableServer(t *testing.T) {
	// Use a closed server so the request fails
	srv := testfetcher.NewMockXKCD(5)
	srv.Close() // close immediately
	f := NewFetcher(srv.URL, 5)
	_, err := f.LastComicID(context.Background())
	assert.Error(t, err)
}

func TestFetcher_Comics(t *testing.T) {
	ctx := context.Background()
	_, f := newMock(t, 20)
	id, err := f.LastComicID(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 20, id)

	ids, comics := f.Comics(ctx, id)
	for i := 1; i <= id; i++ {
		ids <- i
	}

	var result int
	for range comics {
		result++
	}
	assert.Equal(t, id, result)
}

func TestFetcher_Comics_SmallLimit(t *testing.T) {
	ctx := context.Background()
	limit := 3
	_, f := newMock(t, 20)

	ids, comics := f.Comics(ctx, limit)
	for i := 1; i <= limit; i++ {
		ids <- i
	}

	var result int
	for range comics {
		result++
	}
	assert.Equal(t, limit, result)
}

func TestFetcher_Comics_SkipsNotFound(t *testing.T) {
	ctx := context.Background()
	lastID := 5
	_, f := newMock(t, lastID)

	// Send IDs beyond lastID — mock returns 404, which the fetcher skips
	ids, comics := f.Comics(ctx, lastID)
	for i := lastID + 1; i <= lastID*2; i++ {
		ids <- i
	}

	var result int
	for range comics {
		result++
	}
	// All beyond lastID return 404 and are skipped
	assert.Equal(t, 0, result)
}

func TestFetcher_Comics_SingleComic(t *testing.T) {
	ctx := context.Background()
	_, f := newMock(t, 10)

	ids, comics := f.Comics(ctx, 1)
	ids <- 1

	var result int
	for c := range comics {
		result++
		assert.Equal(t, 1, c.ID)
	}
	assert.Equal(t, 1, result)
}
