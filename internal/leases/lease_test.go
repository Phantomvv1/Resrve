package leases

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/Phantomvv1/Resrve/internal/middleware"
	"github.com/Phantomvv1/Resrve/internal/resources"
	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(middleware.ResourceGetterMiddleware)

	r.POST("/leases/:id", AcquireLease)
	r.PUT("/leases/:id/renew", RenewLease)
	r.DELETE("/leases/:id", ReleaseLease)

	return r
}

func resetState() {
	for _, r := range resources.AllResources() {
		r.Lock()
		r.State = resources.StateFree
		r.UpdateLease(nil)
		r.Unlock()
	}
}

func TestAcquireLeaseSuccess(t *testing.T) {
	resetState()
	r := setupRouter()

	req := httptest.NewRequest(http.MethodPost, "/leases/abcd", nil)
	req.RemoteAddr = "1.1.1.1:1234"

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAcquireLeaseConflict(t *testing.T) {
	resetState()
	r := setupRouter()

	req1 := httptest.NewRequest(http.MethodPost, "/leases/abcd", nil)
	req1.RemoteAddr = "1.1.1.1:1234"

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	req2 := httptest.NewRequest(http.MethodPost, "/leases/abcd", nil)
	req2.RemoteAddr = "2.2.2.2:1234"

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Fatalf("expected 409 conflict, got %d", w2.Code)
	}
}

func TestRetryAcquireSameClient(t *testing.T) {
	resetState()
	r := setupRouter()

	req := httptest.NewRequest(http.MethodPost, "/leases/abcd", nil)
	req.RemoteAddr = "1.1.1.1:1234"

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req)

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req)

	if w2.Code != http.StatusConflict {
		t.Fatalf("expected conflict on retry, got %d", w2.Code)
	}
}

func TestAcquireAfterExpiration(t *testing.T) {
	resetState()
	r := setupRouter()

	req1 := httptest.NewRequest(http.MethodPost, "/leases/abcd", nil)
	req1.RemoteAddr = "1.1.1.1:1234"

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	time.Sleep(11 * time.Second)

	req2 := httptest.NewRequest(http.MethodPost, "/leases/abcd", nil)
	req2.RemoteAddr = "2.2.2.2:1234"

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected success after expiration, got %d", w2.Code)
	}
}

func TestRenewAfterExpirationFails(t *testing.T) {
	resetState()
	r := setupRouter()

	req1 := httptest.NewRequest(http.MethodPost, "/leases/abcd", nil)
	req1.RemoteAddr = "1.1.1.1:1234"

	r.ServeHTTP(httptest.NewRecorder(), req1)

	time.Sleep(11 * time.Second)

	reqRenew := httptest.NewRequest(http.MethodPut, "/leases/abcd/renew", nil)
	reqRenew.RemoteAddr = "1.1.1.1:1234"

	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqRenew)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %d", w.Code)
	}
}

func TestRetryRenew(t *testing.T) {
	resetState()
	r := setupRouter()

	req1 := httptest.NewRequest(http.MethodPost, "/leases/abcd", nil)
	req1.RemoteAddr = "1.1.1.1:1234"

	r.ServeHTTP(httptest.NewRecorder(), req1)

	reqRenew := httptest.NewRequest(http.MethodPut, "/leases/abcd/renew", nil)
	reqRenew.RemoteAddr = "1.1.1.1:1234"

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, reqRenew)

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, reqRenew)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected retry renew to succeed, got %d", w2.Code)
	}
}

func TestConcurrentAcquire(t *testing.T) {
	resetState()
	r := setupRouter()

	var wg sync.WaitGroup
	successCount := 0
	mu := sync.Mutex{}

	for i := range 10 {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			req := httptest.NewRequest(http.MethodPost, "/leases/abcd", nil)
			req.RemoteAddr = "client"

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	if successCount != 1 {
		t.Fatalf("expected only 1 success, got %d", successCount)
	}
}

func TestDelayedAcquire(t *testing.T) {
	resetState()
	r := setupRouter()

	req1 := httptest.NewRequest(http.MethodPost, "/leases/abcd", nil)
	req1.RemoteAddr = "1.1.1.1:1234"

	r.ServeHTTP(httptest.NewRecorder(), req1)

	time.Sleep(11 * time.Second)

	req2 := httptest.NewRequest(http.MethodPost, "/leases/abcd", nil)
	req2.RemoteAddr = "2.2.2.2:5678"

	r.ServeHTTP(httptest.NewRecorder(), req2)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req1)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected conflict from stale client, got %d", w.Code)
	}
}

func TestReleaseByNonOwnerFails(t *testing.T) {
	resetState()
	r := setupRouter()

	req1 := httptest.NewRequest(http.MethodPost, "/leases/abcd", nil)
	req1.RemoteAddr = "1.1.1.1:1234"

	r.ServeHTTP(httptest.NewRecorder(), req1)

	reqRelease := httptest.NewRequest(http.MethodDelete, "/leases/abcd", nil)
	reqRelease.RemoteAddr = "2.2.2.2:5678"

	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqRelease)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %d", w.Code)
	}
}
