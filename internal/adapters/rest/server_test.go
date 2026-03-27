package rest

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"yadro-go-course/config"
	xkcdmock "yadro-go-course/test/fetcher"
)

func TestServer_SetupRoutes(t *testing.T) {
	srv := NewServer()
	srv.SetupRoutes(&config.Config{
		Server: config.Server{
			RateLimit:        1,
			ConcurrencyLimit: 1,
			TokenMaxTime:     time.Minute,
			DeleteEvery:      time.Minute,
		},
	}, context.Background())
}

func TestServer_SetupServer(t *testing.T) {
	srv := NewServer()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := &config.Config{
		Fetcher: config.Fetcher{UpdateSpec: "@hourly"},
		Server:  config.Server{Port: 0},
	}
	err := srv.SetupServer(cfg, ctx)
	assert.NoError(t, err)
	assert.NotNil(t, srv.cron)
	assert.Equal(t, "0.0.0.0:0", srv.Srv.Addr)
}

func TestServer_SetupServer_InvalidCron(t *testing.T) {
	srv := NewServer()
	cfg := &config.Config{
		Fetcher: config.Fetcher{UpdateSpec: "not a valid cron spec !@#"},
	}
	err := srv.SetupServer(cfg, context.Background())
	assert.Error(t, err)
}

func TestServer_Stop(t *testing.T) {
	srv := NewServer()
	// Shutdown on a server that was never started returns nil
	err := srv.Stop(context.Background())
	assert.NoError(t, err)
}

func TestNewServer_NotNil(t *testing.T) {
	srv := NewServer()
	assert.NotNil(t, srv)
}

func TestServer_SetupServices(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	stopWordsPath := filepath.Join(tmpDir, "stopwords.txt")

	err := os.WriteFile(stopWordsPath, []byte("the an a"), 0o644)
	require.NoError(t, err)

	mock := xkcdmock.NewMockXKCD(2)
	t.Cleanup(mock.Close)

	cfg := &config.Config{
		DB:      config.DB{Url: dbPath},
		Search:  config.Search{StopWordsFile: stopWordsPath},
		Fetcher: config.Fetcher{SourceURL: mock.URL, Parallel: 2},
	}

	srv := NewServer()
	err = srv.SetupServices(cfg, context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, srv.SearchSrv)
	assert.NotNil(t, srv.ComicSrv)
	assert.NotNil(t, srv.UserSrv)
	assert.NotNil(t, srv.FetcherSrv)
}

func TestServer_SetupServices_BadStopWordsFile(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		DB:     config.DB{Url: filepath.Join(tmpDir, "test.db")},
		Search: config.Search{StopWordsFile: "/nonexistent/stopwords.txt"},
	}
	srv := NewServer()
	err := srv.SetupServices(cfg, context.Background())
	assert.Error(t, err)
}

func TestServer_Start(t *testing.T) {
	srv := NewServer()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := &config.Config{
		Fetcher: config.Fetcher{UpdateSpec: "@hourly"},
		Server:  config.Server{Port: 0},
	}
	require.NoError(t, srv.SetupServer(cfg, ctx))

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	ln.Close()

	srv.Srv.Addr = addr

	errCh := make(chan error, 1)
	go func() { errCh <- srv.Start() }()

	time.Sleep(20 * time.Millisecond)
	require.NoError(t, srv.Stop(context.Background()))

	select {
	case err := <-errCh:
		assert.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return after Stop")
	}
}
