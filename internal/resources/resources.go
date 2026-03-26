package resources

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	StateFree = iota
	StateTaken
)

var ErrNoSuchResource = errors.New("Error: there is no such resource available")
var ErrMakingLeaseId = errors.New("Error: unable to make an id for the new lease")

type Resource struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	State int    `json:"state"`
	lease *Lease
	sync.Mutex
}

func NewResource(id string, name string) *Resource {
	return &Resource{
		ID:    id,
		Name:  name,
		State: StateFree,
		lease: nil,
		Mutex: sync.Mutex{},
	}
}

func (r *Resource) Lease() *Lease {
	return r.lease
}

func (r *Resource) UpdateLease(lease *Lease) {
	r.lease = lease
}

type Lease struct {
	ID           string    `json:"id"`
	Holder       string    `json:"holder"`
	FencingToken int       `json:"fencing_token"`
	TTL          time.Time `json:"ttl"`
}

func NewLease(holder string, fencingToken int) (*Lease, error) {
	idBytes := make([]byte, 32)
	_, err := rand.Read(idBytes)
	if err != nil {
		return nil, ErrMakingLeaseId
	}

	id := base64.StdEncoding.EncodeToString(idBytes)

	return &Lease{
		ID:           id,
		Holder:       holder,
		FencingToken: fencingToken,
		TTL:          time.Now().UTC().Add(time.Second * 10),
	}, nil
}

var availableResources = []*Resource{NewResource("abcd", "Seat 1"), NewResource("efgh", "Seat 2"), NewResource("ijkl", "Seat 3")}

func AllResources() []*Resource {
	return availableResources
}

func GetResourceById(id string) (*Resource, error) {
	for _, resource := range availableResources {
		if resource.ID == id {
			return resource, nil
		}
	}

	return nil, ErrNoSuchResource
}

func GetResource(c *gin.Context) {
	resourceAny, _ := c.Get("resource")
	resource := resourceAny.(*Resource)

	resource.Lock()
	if resource.lease != nil && time.Now().UTC().After(resource.lease.TTL) {
		resource.State = StateFree
		resource.lease = nil
	}

	resource.Unlock()

	c.JSON(http.StatusOK, gin.H{"result": resource})
}

func GetAllResources(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"result": availableResources})
}
