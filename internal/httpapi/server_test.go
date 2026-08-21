package httpapi

import (
	"example.com/task149/tsnsched/internal/metrics"
	"example.com/task149/tsnsched/internal/service"
	"example.com/task149/tsnsched/internal/store"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestHealthRoute(t *testing.T) {
	s, e := store.Open(filepath.Join(t.TempDir(), "s.db"))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	r := httptest.NewRecorder()
	New(service.New(s), metrics.New()).Handler().ServeHTTP(r, httptest.NewRequest("GET", "/healthz", nil))
	if r.Code != 200 {
		t.Fatalf("status %d", r.Code)
	}
}
