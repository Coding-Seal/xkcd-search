package integration_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"yadro-go-course/config"
	"yadro-go-course/db"
	comicrepo "yadro-go-course/internal/adapters/repos/comic"
	fetcherrepo "yadro-go-course/internal/adapters/repos/fetcher"
	searchrepo "yadro-go-course/internal/adapters/repos/search"
	userrepo "yadro-go-course/internal/adapters/repos/user"
	"yadro-go-course/internal/adapters/rest"
	"yadro-go-course/internal/core/models"
	"yadro-go-course/internal/core/ports"
	"yadro-go-course/internal/core/services"
	xkcdmock "yadro-go-course/test/fetcher"
	"yadro-go-course/pkg/words"
)

type testEnv struct {
	srv        *httptest.Server
	adminToken string
	userToken  string
	comicRepo  ports.ComicsRepo
	index      *searchrepo.Index
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := &config.Config{
		DB: config.DB{
			Url: dbPath,
		},
		Server: config.Server{
			RateLimit:        100,
			ConcurrencyLimit: 10,
			TokenMaxTime:     time.Hour,
			DeleteEvery:      time.Minute,
		},
		Fetcher: config.Fetcher{
			SourceURL:  "http://localhost",
			Parallel:   5,
			UpdateSpec: "@hourly",
		},
	}

	conn, err := sql.Open("sqlite3", cfg.DB.Url)
	require.NoError(t, err, "open sqlite db")
	t.Cleanup(func() { conn.Close() })

	err = db.MigrateUp(conn, ctx)
	require.NoError(t, err, "run migrations")

	comicRepo := comicrepo.NewSqliteStore(conn)
	userRepo := userrepo.NewSqliteRepo(conn)

	mockXKCD := xkcdmock.NewMockXKCD(5)
	t.Cleanup(mockXKCD.Close)

	fetcherRepo := fetcherrepo.NewFetcher(mockXKCD.URL, cfg.Fetcher.Parallel)

	stemmer := words.NewStemmer(nil)
	index := searchrepo.NewIndex(stemmer)

	searchSrv := services.NewSearch(index, comicRepo)
	fetcherSrv := services.NewFetcher(fetcherRepo, comicRepo, index)
	userSrv := services.NewUserService(userRepo)
	comicSrv := services.NewComicService(comicRepo)

	// Create admin user
	adminHash, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	require.NoError(t, err, "hash admin password")
	adminUser := &models.User{
		Login:    "admin",
		Password: adminHash,
		IsAdmin:  true,
	}
	err = userRepo.AddUser(ctx, adminUser)
	require.NoError(t, err, "add admin user")

	// Create regular user
	userHash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	require.NoError(t, err, "hash user password")
	regularUser := &models.User{
		Login:    "user",
		Password: userHash,
		IsAdmin:  false,
	}
	err = userRepo.AddUser(ctx, regularUser)
	require.NoError(t, err, "add regular user")

	handler := rest.Api(fetcherSrv, searchSrv, userSrv, comicSrv, cfg, ctx)
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	// Obtain admin token
	adminToken := loginAndGetToken(t, srv.URL, "admin", "admin")

	// Obtain regular user token
	userToken := loginAndGetToken(t, srv.URL, "user", "password")

	return &testEnv{
		srv:        srv,
		adminToken: adminToken,
		userToken:  userToken,
		comicRepo:  comicRepo,
		index:      index,
	}
}

func loginAndGetToken(t *testing.T, baseURL, login, password string) string {
	t.Helper()

	body, err := json.Marshal(map[string]string{"login": login, "password": password})
	require.NoError(t, err)

	resp, err := http.Post(baseURL+"/login", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "login should succeed for %s", login)

	token := resp.Header.Get("Authorization")
	require.NotEmpty(t, token, "token should be present in Authorization header")

	return token
}

// TestIntegration_Login_Success verifies that valid admin credentials return 200 with a JWT token.
func TestIntegration_Login_Success(t *testing.T) {
	env := setupTestEnv(t)

	body, err := json.Marshal(map[string]string{"login": "admin", "password": "admin"})
	require.NoError(t, err)

	resp, err := http.Post(env.srv.URL+"/login", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	token := resp.Header.Get("Authorization")
	assert.NotEmpty(t, token, "Authorization header should contain JWT token")
}

