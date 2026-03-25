package resources

import (
	"errors"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

const (
	StateFree = iota
	StateTaken
)

var ErrNoSuchResource = errors.New("Error: there is no such resource available")

type Resource struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	State int    `json:"state"`
	Mu    sync.Mutex
}

func NewResource(id string, name string) *Resource {
	return &Resource{
		ID:    id,
		Name:  name,
		State: StateFree,
		Mu:    sync.Mutex{},
	}
}

func (r *Resource) UpdateState(state int) {
	r.Mu.Lock()
	r.State = state
	r.Mu.Unlock()
}

var availableResources = []*Resource{NewResource("abcd", "Seat 1"), NewResource("efgh", "Seat 2"), NewResource("ijkl", "Seat 3")}

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

	c.JSON(http.StatusOK, gin.H{"result": resource})
}

func GetAllResources(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"result": availableResources})
}
