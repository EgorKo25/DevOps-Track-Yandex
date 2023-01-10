package routers

import (
	"context"
	config "github.com/EgorKo25/DevOps-Track-Yandex/internal/configuration"
	"github.com/EgorKo25/DevOps-Track-Yandex/internal/database"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EgorKo25/DevOps-Track-Yandex/internal/hashing"
	"github.com/EgorKo25/DevOps-Track-Yandex/internal/middleware"
	"github.com/EgorKo25/DevOps-Track-Yandex/internal/server/handlers"
	"github.com/EgorKo25/DevOps-Track-Yandex/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRouter(t *testing.T) {

	ctx := context.Background()
	cfg := config.NewServerConfig()
	mem := storage.NewStorage()
	cpr := middleware.NewCompressor()
	hsr := hashing.MewHash(cfg.Key)
	db := database.NewDB(cfg, ctx, mem)
	handler := handlers.NewHandler(mem, cpr, hsr, db, ctx)

	value := storage.Gauge(123)
	metric := storage.Metric{
		ID:    "Alloc",
		MType: "gauge",
		Delta: nil,
		Value: &value,
	}

	r := NewRouter(handler)
	ts := httptest.NewServer(r)
	defer ts.Close()

	mem.SetStat(&metric)

	statusCode, body := testRequest(t, ts, "GET", "/")
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "> Alloc:  123.000000\n", body)

	statusCode, body = testRequest(t, ts, "GET", "/value/gauge/Alloc")
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "123\n", body)

	statusCode, body = testRequest(t, ts, "POST", "/update/gauge/Alloc/123")
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "", body)

}
func testRequest(t *testing.T, ts *httptest.Server, method, path string) (int, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp.StatusCode, string(respBody)
}
