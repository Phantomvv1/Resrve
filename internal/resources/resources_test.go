package resources

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func resetState() {
	for _, r := range availableResources {
		r.Lock()
		r.State = StateFree
		r.lease = nil
		r.Unlock()
	}
}

func setupGin() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func TestNewResource(t *testing.T) {
	r := NewResource("id1", "Seat A")

	if r.ID != "id1" || r.Name != "Seat A" {
		t.Fatalf("resource fields not set correctly")
	}

	if r.State != StateFree {
		t.Fatalf("expected StateFree, got %d", r.State)
	}

	if r.Lease() != nil {
		t.Fatalf("expected no lease initially")
	}
}

func TestNewLease(t *testing.T) {
	lease, err := NewLease("client", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if lease.Holder != "client" {
		t.Fatalf("wrong holder")
	}

	if lease.FencingToken != 1 {
		t.Fatalf("wrong fencing token")
	}

	if lease.ID == "" {
		t.Fatalf("expected non-empty ID")
	}

	if time.Now().UTC().After(lease.TTL) {
		t.Fatalf("TTL should be in the future")
	}
}

func TestGetResourceByIdSuccess(t *testing.T) {
	resetState()

	r, err := GetResourceById("abcd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if r.ID != "abcd" {
		t.Fatalf("wrong resource returned")
	}
}

func TestGetResourceByIdFail(t *testing.T) {
	resetState()

	_, err := GetResourceById("does-not-exist")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestLeaseAttachAndRetrieve(t *testing.T) {
	resetState()

	r := availableResources[0]

	lease, _ := NewLease("client", 0)

	r.Lock()
	r.UpdateLease(lease)
	r.Unlock()

	if r.Lease() == nil {
		t.Fatalf("lease should be attached")
	}
}

func TestLeaseExpirationOnGetResource(t *testing.T) {
	resetState()

	r := availableResources[0]

	expiredLease := &Lease{
		ID:           "test",
		Holder:       "client",
		FencingToken: 0,
		TTL:          time.Now().UTC().Add(-1 * time.Second),
	}

	r.Lock()
	r.lease = expiredLease
	r.State = StateTaken
	r.Unlock()

	router := setupGin()

	router.GET("/resource", func(c *gin.Context) {
		c.Set("resource", r)
		GetResource(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/resource", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if r.State != StateFree {
		t.Fatalf("expected resource to be freed after expiration")
	}

	if r.lease != nil {
		t.Fatalf("expected lease to be cleared")
	}
}

func TestLeaseNotExpired(t *testing.T) {
	resetState()

	r := availableResources[0]

	validLease := &Lease{
		ID:           "test",
		Holder:       "client",
		FencingToken: 0,
		TTL:          time.Now().UTC().Add(10 * time.Second),
	}

	r.Lock()
	r.lease = validLease
	r.State = StateTaken
	r.Unlock()

	router := setupGin()

	router.GET("/resource", func(c *gin.Context) {
		c.Set("resource", r)
		GetResource(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/resource", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if r.State != StateTaken {
		t.Fatalf("resource should remain taken")
	}

	if r.lease == nil {
		t.Fatalf("lease should still exist")
	}
}

func TestGetAllResources(t *testing.T) {
	resetState()

	router := setupGin()
	router.GET("/resources", GetAllResources)

	req := httptest.NewRequest(http.MethodGet, "/resources", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
