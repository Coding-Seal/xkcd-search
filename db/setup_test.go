package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"yadro-go-course/config"
)

func newInMemoryCfg() *config.Config {
	cfg := &config.Config{}
	cfg.DB.Url = ":memory:"
	return cfg
}

func TestConnect(t *testing.T) {
	db, err := Connect(newInMemoryCfg())
	assert.NoError(t, err)
	assert.NotNil(t, db)
}

func TestMigrateUp(t *testing.T) {
	db, err := Connect(newInMemoryCfg())
	assert.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, MigrateDown(db, context.Background())) })
	assert.NoError(t, MigrateUp(db, context.Background()))
}

func TestMigrateDown(t *testing.T) {
	db, err := Connect(newInMemoryCfg())
	assert.NoError(t, err)
	assert.NoError(t, MigrateUp(db, context.Background()))
	assert.NoError(t, MigrateDown(db, context.Background()))
}

func TestConnect_InMemory(t *testing.T) {
	db, err := Connect(newInMemoryCfg())
	assert.NoError(t, err)
	assert.NotNil(t, db)
}

func TestConnect_Ping(t *testing.T) {
	db, err := Connect(newInMemoryCfg())
	assert.NoError(t, err)
	assert.NoError(t, db.Ping())
}

func TestMigrateUp_Idempotent(t *testing.T) {
	db, err := Connect(newInMemoryCfg())
	assert.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, MigrateDown(db, context.Background())) })
	assert.NoError(t, MigrateUp(db, context.Background()))
	// Running twice should not return error (ErrNoChange is swallowed)
	assert.NoError(t, MigrateUp(db, context.Background()))
}

func TestMigrateDown_Idempotent(t *testing.T) {
	db, err := Connect(newInMemoryCfg())
	assert.NoError(t, err)
	assert.NoError(t, MigrateDown(db, context.Background()))
	// Second down when already empty is fine
	assert.NoError(t, MigrateDown(db, context.Background()))
}

func TestConnect_EmptyURL(t *testing.T) {
	cfg := &config.Config{}
	cfg.DB.Url = ""
	// Empty URL — driver should either error or return a usable connection.
	// We only assert that the call does not panic.
	db, err := Connect(cfg)
	if err == nil {
		assert.NotNil(t, db)
	}
}
