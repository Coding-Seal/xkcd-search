package web

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"yadro-go-course/web/rest"
)

func TestRoutes_NotNil(t *testing.T) {
	client := rest.NewClient("http://localhost:8090")
	handler := Routes(client)
	assert.NotNil(t, handler)
}

func TestRoutes_DifferentBaseURLs(t *testing.T) {
	for _, url := range []string{"http://localhost:8080", "http://127.0.0.1:9000", "http://api.example.com"} {
		client := rest.NewClient(url)
		handler := Routes(client)
		assert.NotNil(t, handler, "Routes should return handler for URL %s", url)
	}
}
