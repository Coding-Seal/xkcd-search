package integration_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

// TestIntegration_Login_WrongPassword verifies that wrong password returns 403.
func TestIntegration_Login_WrongPassword(t *testing.T) {
	env := setupTestEnv(t)

	body, err := json.Marshal(map[string]string{"login": "admin", "password": "wrongpassword"})
	require.NoError(t, err)

	resp, err := http.Post(env.srv.URL+"/login", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

// TestIntegration_Login_UserNotFound verifies that unknown user returns 404.
func TestIntegration_Login_UserNotFound(t *testing.T) {
	env := setupTestEnv(t)

	body, err := json.Marshal(map[string]string{"login": "unknownuser", "password": "somepassword"})
	require.NoError(t, err)

	resp, err := http.Post(env.srv.URL+"/login", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestIntegration_Login_MissingCredentials verifies that empty login or password returns 400.
func TestIntegration_Login_MissingCredentials(t *testing.T) {
	env := setupTestEnv(t)

	testCases := []struct {
		name  string
		login string
		pass  string
	}{
		{"empty login", "", "password"},
		{"empty password", "admin", ""},
		{"both empty", "", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(map[string]string{"login": tc.login, "password": tc.pass})
			require.NoError(t, err)

			resp, err := http.Post(env.srv.URL+"/login", "application/json", bytes.NewReader(body))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}

// TestIntegration_Login_InvalidJSON verifies that malformed JSON returns 400.
func TestIntegration_Login_InvalidJSON(t *testing.T) {
	env := setupTestEnv(t)

	resp, err := http.Post(env.srv.URL+"/login", "application/json", strings.NewReader("{not valid json"))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestIntegration_Search_EmptyQuery verifies that an empty search query returns 404.
func TestIntegration_Search_EmptyQuery(t *testing.T) {
	env := setupTestEnv(t)

	resp, err := http.Get(env.srv.URL + "/pics?search=")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestIntegration_Search_Found verifies that after indexing a comic, searching for it returns 200 with results.
func TestIntegration_Search_Found(t *testing.T) {
	env := setupTestEnv(t)

	ctx := context.Background()

	comic := models.Comic{
		ID:            42,
		Title:         "testcomicstory",
		Transcription: "testcomicstory",
	}
	require.NoError(t, env.comicRepo.Store(ctx, comic), "store test comic")

	err := env.index.Build(ctx, env.comicRepo)
	require.NoError(t, err, "build index")

	resp, err := http.Get(env.srv.URL + "/pics?search=testcomicstory")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var results []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&results)
	require.NoError(t, err, "decode search results")
	assert.NotEmpty(t, results, "search should return at least one result")
}

// TestIntegration_Update_NoToken verifies that POST /update without a token returns 401.
func TestIntegration_Update_NoToken(t *testing.T) {
	env := setupTestEnv(t)

	resp, err := http.Post(env.srv.URL+"/update", "application/json", nil)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestIntegration_Update_WithAdminToken verifies that POST /update with an admin token returns 200.
func TestIntegration_Update_WithAdminToken(t *testing.T) {
	env := setupTestEnv(t)

	req, err := http.NewRequest(http.MethodPost, env.srv.URL+"/update", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", env.adminToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestIntegration_Update_WithNonAdminToken verifies that POST /update with a non-admin token returns 401.
func TestIntegration_Update_WithNonAdminToken(t *testing.T) {
	env := setupTestEnv(t)

	req, err := http.NewRequest(http.MethodPost, env.srv.URL+"/update", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", env.userToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestIntegration_Comic_InvalidID verifies that GET /comic/abc returns 400.
func TestIntegration_Comic_InvalidID(t *testing.T) {
	env := setupTestEnv(t)

	resp, err := http.Get(env.srv.URL + "/comic/abc")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestIntegration_Concurrent_Requests verifies that multiple concurrent requests all complete without server crash.
func TestIntegration_Concurrent_Requests(t *testing.T) {
	env := setupTestEnv(t)

	const totalRequests = 20
	var wg sync.WaitGroup
	statusCodes := make([]int, totalRequests)

	for i := 0; i < totalRequests; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := http.Get(fmt.Sprintf("%s/pics?search=test%d", env.srv.URL, i))
			if err != nil {
				return
			}
			defer resp.Body.Close()
			statusCodes[i] = resp.StatusCode
		}()
	}
	wg.Wait()

	for _, code := range statusCodes {
		// Each request should return either 200 (found) or 404 (not found)
		assert.Contains(t, []int{http.StatusOK, http.StatusNotFound}, code)
	}
}

// TestIntegration_Comic_Found verifies that GET /comic/{id} returns 200 for an existing comic.
func TestIntegration_Comic_Found(t *testing.T) {
	env := setupTestEnv(t)

	ctx := context.Background()
	comic := models.Comic{ID: 7, Title: "Found Comic", ImgURL: "http://example.com/7.png"}
	require.NoError(t, env.comicRepo.Store(ctx, comic))

	resp, err := http.Get(env.srv.URL + "/comic/7")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, float64(7), result["id"])
}

// TestIntegration_Comic_UnknownID verifies that GET /comic/{id} for a non-existing comic returns 500 (handler returns ErrInternal for all repo errors).
func TestIntegration_Comic_UnknownID(t *testing.T) {
	env := setupTestEnv(t)

	resp, err := http.Get(env.srv.URL + "/comic/99999")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// TestIntegration_Search_MultipleResults verifies that storing multiple comics and indexing them returns multiple results.
func TestIntegration_Search_MultipleResults(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	for i, title := range []string{"uniquexyzword story one", "uniquexyzword adventure two", "uniquexyzword tale three"} {
		c := models.Comic{ID: 100 + i, Title: title, Transcription: title}
		require.NoError(t, env.comicRepo.Store(ctx, c))
	}
	require.NoError(t, env.index.Build(ctx, env.comicRepo))

	resp, err := http.Get(env.srv.URL + "/pics?search=uniquexyzword")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var results []map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&results))
	assert.GreaterOrEqual(t, len(results), 2, "should return multiple results")
}

// TestIntegration_Search_Pics_NoParam verifies that GET /pics without any search parameter returns 404.
func TestIntegration_Search_Pics_NoParam(t *testing.T) {
	env := setupTestEnv(t)

	resp, err := http.Get(env.srv.URL + "/pics")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestIntegration_Auth_GarbageToken verifies that a garbage Authorization header returns 401.
func TestIntegration_Auth_GarbageToken(t *testing.T) {
	env := setupTestEnv(t)

	req, err := http.NewRequest(http.MethodPost, env.srv.URL+"/update", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "this-is-garbage-not-a-jwt")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestIntegration_Auth_WrongSignatureJWT verifies that a JWT signed with the wrong key returns 401.
func TestIntegration_Auth_WrongSignatureJWT(t *testing.T) {
	env := setupTestEnv(t)

	claims := jwt.MapClaims{
		"user_id":  1,
		"is_admin": true,
		"exp":      time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString([]byte("wrong-secret-key-for-testing"))
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, env.srv.URL+"/update", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", tokenStr)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestIntegration_Update_FetchesComics verifies that POST /update with admin token causes comics to be fetchable.
func TestIntegration_Update_FetchesComics(t *testing.T) {
	env := setupTestEnv(t)

	req, err := http.NewRequest(http.MethodPost, env.srv.URL+"/update", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", env.adminToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// After update, comic #1 from the mock XKCD server should be in the DB
	resp2, err := http.Get(env.srv.URL + "/comic/1")
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusOK, resp2.StatusCode)
}

// TestIntegration_Search_StopWordsOnly verifies that a query consisting only of filtered words returns 404.
func TestIntegration_Search_StopWordsOnly(t *testing.T) {
	env := setupTestEnv(t)

	// "about", "above", "after" are stop words — stemmer filters them → empty search → 404
	resp, err := http.Get(env.srv.URL + "/pics?search=about+above+after")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestIntegration_Search_Stemming verifies that search matches stemmed word forms (e.g. "running" matches "run").
func TestIntegration_Search_Stemming(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	comic := models.Comic{ID: 200, Title: "running", Transcription: "running"}
	require.NoError(t, env.comicRepo.Store(ctx, comic))
	require.NoError(t, env.index.Build(ctx, env.comicRepo))

	// "runs" should stem to the same root as "running" and find the comic
	resp, err := http.Get(env.srv.URL + "/pics?search=runs")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var results []map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&results))
	assert.NotEmpty(t, results)
}

// TestIntegration_Search_HigherScoreComicRankedFirst verifies that a comic matching more query words
// appears before a comic matching fewer words.
func TestIntegration_Search_HigherScoreComicRankedFirst(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	// comic 301 matches only one keyword; comic 302 matches both
	require.NoError(t, env.comicRepo.Store(ctx, models.Comic{ID: 301, Title: "galaxybrain", Transcription: "galaxybrain"}))
	require.NoError(t, env.comicRepo.Store(ctx, models.Comic{ID: 302, Title: "galaxybrain rocketship", Transcription: "galaxybrain rocketship"}))
	require.NoError(t, env.index.Build(ctx, env.comicRepo))

	resp, err := http.Get(env.srv.URL + "/pics?search=galaxybrain+rocketship")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var results []map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&results))
	require.GreaterOrEqual(t, len(results), 2, "both comics should be returned")

	// comic 302 matches both words → higher score → must appear first
	firstID := int(results[0]["id"].(float64))
	assert.Equal(t, 302, firstID, "comic matching more terms should rank first")
}

// TestIntegration_Search_UpdateThenSearch verifies the full pipeline: update fetches comics
// and they become searchable immediately after.
func TestIntegration_Search_UpdateThenSearch(t *testing.T) {
	env := setupTestEnv(t)

	// Trigger update with admin token
	req, err := http.NewRequest(http.MethodPost, env.srv.URL+"/update", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", env.adminToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// The mock XKCD server serves comics with title "testStr" — search for that
	resp2, err := http.Get(env.srv.URL + "/pics?search=testStr")
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	var results []map[string]interface{}
	require.NoError(t, json.NewDecoder(resp2.Body).Decode(&results))
	assert.NotEmpty(t, results, "comics fetched by update should be searchable")
}

// TestIntegration_Search_ReturnsCorrectFields verifies the JSON shape of search results.
func TestIntegration_Search_ReturnsCorrectFields(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	comic := models.Comic{ID: 400, Title: "fieldcheckcomic", ImgURL: "http://example.com/400.png"}
	require.NoError(t, env.comicRepo.Store(ctx, comic))
	require.NoError(t, env.index.Build(ctx, env.comicRepo))

	resp, err := http.Get(env.srv.URL + "/pics?search=fieldcheckcomic")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var results []map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&results))
	require.NotEmpty(t, results)

	first := results[0]
	assert.Contains(t, first, "id", "result must have 'id' field")
	assert.Contains(t, first, "img", "result must have 'img' field")
}

// TestIntegration_Comic_ReturnedFieldsCorrect verifies GET /comic/{id} returns expected JSON fields.
func TestIntegration_Comic_ReturnedFieldsCorrect(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	comic := models.Comic{ID: 500, Title: "FieldComic", ImgURL: "http://example.com/500.png"}
	require.NoError(t, env.comicRepo.Store(ctx, comic))

	resp, err := http.Get(env.srv.URL + "/comic/500")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))

	assert.Equal(t, float64(500), result["id"])
	assert.Equal(t, "http://example.com/500.png", result["img"])
}

// TestIntegration_Update_Idempotent verifies that calling update twice does not duplicate comics in search results.
func TestIntegration_Update_Idempotent(t *testing.T) {
	env := setupTestEnv(t)

	doUpdate := func() {
		req, err := http.NewRequest(http.MethodPost, env.srv.URL+"/update", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", env.adminToken)
		resp, err := (&http.Client{}).Do(req)
		require.NoError(t, err)
		resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	}

	doUpdate()
	doUpdate()

	resp, err := http.Get(env.srv.URL + "/pics?search=testStr")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var results []map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&results))

	// Collect IDs and verify no duplicates
	seen := map[float64]bool{}
	for _, r := range results {
		id := r["id"].(float64)
		assert.False(t, seen[id], "duplicate comic id %v in results", id)
		seen[id] = true
	}
}

// TestIntegration_Search_MultiWordQuery verifies that a multi-word query finds a comic containing all those words.
func TestIntegration_Search_MultiWordQuery(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	require.NoError(t, env.comicRepo.Store(ctx, models.Comic{
		ID:            600,
		Title:         "zirconium telescope discovery",
		Transcription: "zirconium telescope discovery",
	}))
	require.NoError(t, env.index.Build(ctx, env.comicRepo))

	resp, err := http.Get(env.srv.URL + "/pics?search=zirconium+telescope")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var results []map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&results))
	assert.NotEmpty(t, results)
}
