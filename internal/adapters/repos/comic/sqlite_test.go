package comic

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"yadro-go-course/config"
	"yadro-go-course/db"
	"yadro-go-course/internal/core/models"
	"yadro-go-course/internal/core/ports"
)

func newTestStore(t *testing.T) *SqliteStore {
	t.Helper()
	cfg := &config.Config{DB: config.DB{Url: ":memory:"}}
	conn, err := db.Connect(cfg)
	assert.NoError(t, err)
	ctx := context.Background()
	assert.NoError(t, db.MigrateUp(conn, ctx))
	t.Cleanup(func() { assert.NoError(t, db.MigrateDown(conn, ctx)) })
	return NewSqliteStore(conn)
}

func TestSqliteStore_Comic_NotFound(t *testing.T) {
	store := newTestStore(t)
	_, err := store.Comic(context.Background(), 999)
	assert.ErrorIs(t, err, ports.ErrNotFound)
}

func TestSqliteStore_Comic_ZeroID(t *testing.T) {
	store := newTestStore(t)
	_, err := store.Comic(context.Background(), 0)
	assert.ErrorIs(t, err, ports.ErrNotFound)
}

func TestSqliteStore_Comic(t *testing.T) {
	testComic := models.Comic{ID: 1, Title: "test"}
	ctx := context.Background()
	store := newTestStore(t)

	_, err := store.Comic(ctx, 0)
	assert.ErrorIs(t, err, ports.ErrNotFound)

	assert.NoError(t, store.Store(ctx, testComic))
	comic, err := store.Comic(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, testComic, comic)

	assert.ErrorIs(t, store.Store(ctx, testComic), ports.ErrInternal)
}

func TestSqliteStore_Comics_Empty(t *testing.T) {
	store := newTestStore(t)
	comics, err := store.ComicsAll(context.Background())
	assert.NoError(t, err)
	assert.Empty(t, comics)
}

func TestSqliteStore_Comics(t *testing.T) {
	var testComics []models.Comic
	ctx := context.Background()
	store := newTestStore(t)

	for i := 1; i <= 5; i++ {
		comic := models.Comic{ID: i, Title: "test"}
		testComics = append(testComics, comic)
		assert.NoError(t, store.Store(ctx, comic))
	}

	comics, err := store.ComicsAll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, testComics, comics)
}

func TestSqliteStore_Store_PreservesFields(t *testing.T) {
	testComic := models.Comic{
		ID:               7,
		Title:            "The Title",
		SafeTitle:        "safe",
		ImgURL:           "https://example.com/img.png",
		Transcription:    "transcript text",
		AltTranscription: "alt text",
		News:             "news text",
		Link:             "https://example.com",
	}
	ctx := context.Background()
	store := newTestStore(t)
	assert.NoError(t, store.Store(ctx, testComic))

	got, err := store.Comic(ctx, testComic.ID)
	assert.NoError(t, err)
	assert.Equal(t, testComic, got)
}

func TestSqliteStore_Store_DuplicateIsError(t *testing.T) {
	comic := models.Comic{ID: 1, Title: "dup"}
	ctx := context.Background()
	store := newTestStore(t)
	assert.NoError(t, store.Store(ctx, comic))
	assert.ErrorIs(t, store.Store(ctx, comic), ports.ErrInternal)
}

func TestSqliteStore_Comics_Order(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	// Insert in reverse order; ComicsAll should return them ordered by id
	for i := 5; i >= 1; i-- {
		assert.NoError(t, store.Store(ctx, models.Comic{ID: i, Title: "t"}))
	}
	comics, err := store.ComicsAll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(comics))
	for i, c := range comics {
		assert.Equal(t, i+1, c.ID)
	}
}
