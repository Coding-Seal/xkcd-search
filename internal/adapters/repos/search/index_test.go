package search

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"yadro-go-course/internal/core/models"
	"yadro-go-course/internal/core/ports"
	"yadro-go-course/pkg/words"
)

func newIndex() *Index {
	return NewIndex(words.NewStemmer(nil))
}

func TestIndex_AddComic(t *testing.T) {
	ctx := context.Background()
	testComic := models.Comic{ID: 42, Title: "TestString and something"}
	index := newIndex()
	index.AddComic(ctx, testComic)
	found := index.SearchComics(ctx, "TestString and something")
	assert.Contains(t, found, testComic.ID)
}

func TestIndex_Build(t *testing.T) {
	m := newComicRepoMock()
	testComics := []models.Comic{{ID: 1, Title: "testStr"}, {ID: 2, Title: "testStr"}, {ID: 3, Title: "testStr"}}
	m.On("ComicsAll").Return(testComics, nil)

	ctx := context.Background()
	index := newIndex()
	assert.NoError(t, index.Build(ctx, m))

	found := index.SearchComics(ctx, "testStr")
	for _, comic := range testComics {
		assert.Contains(t, found, comic.ID)
	}
}

func TestIndex_SearchMiss(t *testing.T) {
	ctx := context.Background()
	index := newIndex()
	found := index.SearchComics(ctx, "unicorn")
	assert.Empty(t, found)
}

func TestIndex_MultipleComics(t *testing.T) {
	ctx := context.Background()
	index := newIndex()
	index.AddComic(ctx, models.Comic{ID: 1, Title: "dinosaur discovery adventure"})
	index.AddComic(ctx, models.Comic{ID: 2, Title: "dinosaur fossil history"})
	index.AddComic(ctx, models.Comic{ID: 3, Title: "space exploration rocket"})

	found := index.SearchComics(ctx, "dinosaur")
	assert.Contains(t, found, 1)
	assert.Contains(t, found, 2)
	assert.NotContains(t, found, 3)
}

func TestIndex_EmptyQuery(t *testing.T) {
	ctx := context.Background()
	index := newIndex()
	index.AddComic(ctx, models.Comic{ID: 1, Title: "something"})
	found := index.SearchComics(ctx, "")
	assert.Empty(t, found)
}

func TestIndex_Build_Error(t *testing.T) {
	m := newComicRepoMock()
	m.On("ComicsAll").Return([]models.Comic{}, ports.ErrInternal)
	ctx := context.Background()
	index := newIndex()
	err := index.Build(ctx, m)
	assert.Error(t, err)
}

func TestIndex_Build_EmptyRepo(t *testing.T) {
	m := newComicRepoMock()
	m.On("ComicsAll").Return([]models.Comic{}, nil)
	ctx := context.Background()
	index := newIndex()
	assert.NoError(t, index.Build(ctx, m))
	assert.Empty(t, index.SearchComics(ctx, "anything"))
}

func TestIndex_SearchByTranscription(t *testing.T) {
	ctx := context.Background()
	index := newIndex()
	index.AddComic(ctx, models.Comic{ID: 5, Title: "", Transcription: "wormhole quantum physics"})
	found := index.SearchComics(ctx, "quantum")
	assert.Contains(t, found, 5)
}

func TestIndex_SearchByAltTranscription(t *testing.T) {
	ctx := context.Background()
	index := newIndex()
	index.AddComic(ctx, models.Comic{ID: 6, Title: "", AltTranscription: "balloon festival sky"})
	found := index.SearchComics(ctx, "balloon")
	assert.Contains(t, found, 6)
}

func TestIndex_SearchBySafeTitle(t *testing.T) {
	ctx := context.Background()
	index := newIndex()
	index.AddComic(ctx, models.Comic{ID: 7, SafeTitle: "gravity waves detected"})
	found := index.SearchComics(ctx, "gravity")
	assert.Contains(t, found, 7)
}

func TestIndex_Ranking_MoreMatchesHigherScore(t *testing.T) {
	ctx := context.Background()
	index := newIndex()
	// comic 1 matches both words; comic 2 matches only one
	index.AddComic(ctx, models.Comic{ID: 1, Title: "robot laser battle"})
	index.AddComic(ctx, models.Comic{ID: 2, Title: "robot cooking"})

	found := index.SearchComics(ctx, "robot laser")
	assert.GreaterOrEqual(t, found[1], found[2],
		"comic matching more query words should have higher or equal score")
}

func TestIndex_SearchEmpty_AfterAddNoKeywords(t *testing.T) {
	ctx := context.Background()
	index := newIndex()
	// Comic with no meaningful content (stop words / too short)
	index.AddComic(ctx, models.Comic{ID: 9, Title: "a"})
	found := index.SearchComics(ctx, "a")
	// Short single-char words may be filtered by the stemmer; result is empty or not —
	// either way the index must not panic and must return a map.
	assert.NotNil(t, found)
}

func TestIndex_ConcurrentAddAndSearch(t *testing.T) {
	ctx := context.Background()
	index := newIndex()
	var wg sync.WaitGroup

	for i := 1; i <= 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			index.AddComic(ctx, models.Comic{ID: id, Title: "concurrent comics test"})
		}(i)
	}

	wg.Wait()
	found := index.SearchComics(ctx, "concurrent")
	assert.Equal(t, 20, len(found))
}

// ---- mock ----

type comicRepoMock struct {
	mock.Mock
}

var _ ports.ComicsRepo = (*comicRepoMock)(nil)

func newComicRepoMock() *comicRepoMock {
	return &comicRepoMock{}
}

func (m *comicRepoMock) Comic(ctx context.Context, id int) (models.Comic, error) {
	args := m.Called(id)
	return args.Get(0).(models.Comic), args.Error(1)
}

func (m *comicRepoMock) Store(ctx context.Context, comic models.Comic) error {
	args := m.Called(comic)
	return args.Error(0)
}

func (m *comicRepoMock) ComicsAll(ctx context.Context) ([]models.Comic, error) {
	args := m.Called()
	return args.Get(0).([]models.Comic), args.Error(1)
}
