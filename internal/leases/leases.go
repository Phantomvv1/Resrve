package leases

import (
	"crypto/rand"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Phantomvv1/Resrve/internal/resources"
	"github.com/gin-gonic/gin"
)

var ErrNoSuchResource = errors.New("Error: there is no such resource available")
var ErrMakingLeaseId = errors.New("Error: unable to make an id for the new lease")

var leaseMap = make(map[*resources.Resource]*Lease)
var muLeaseMap = sync.Mutex{}

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

	return &Lease{
		ID:           string(idBytes),
		Holder:       holder,
		FencingToken: fencingToken,
		TTL:          time.Now().Add(time.Second * 10),
	}, nil
}

func getResourceById(id string) (*resources.Resource, error) {
	for _, resource := range resources.AvailableResources {
		if resource.ID == id {
			return resource, nil
		}
	}

	return nil, ErrNoSuchResource
}

func CreateAndReturnLease(c *gin.Context, resource *resources.Resource, fencingToken int) {
	resultLease, err := NewLease(c.ClientIP(), fencingToken)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	muLeaseMap.Lock()
	leaseMap[resource] = resultLease
	muLeaseMap.Unlock()

	c.JSON(http.StatusOK, gin.H{"result": resultLease})
}

func AcquireLease(c *gin.Context) {
	now := time.Now().UTC()
	id := c.Param("id")

	resource, err := getResourceById(id)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	muLeaseMap.Lock()
	lease, ok := leaseMap[resource]
	muLeaseMap.Unlock()

	if !ok {
		CreateAndReturnLease(c, resource, 0)
		return
	}

	if now.After(lease.TTL) {
		CreateAndReturnLease(c, resource, lease.FencingToken+1)
		return
	} else {
		c.JSON(http.StatusConflict, gin.H{"error": "Error: the resource is unavailable at the moment"})
		return
	}
}

func RenewLease(c *gin.Context) {

}

func ReleaseLease(c *gin.Context) {

}

func GetLease(c *gin.Context) {
	id := c.Param("id")

	resource, err := getResourceById(id)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": resource})
}

func GetAllLeases(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"result": resources.AvailableResources})
}
