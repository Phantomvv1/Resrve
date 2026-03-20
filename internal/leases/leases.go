package leases

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Phantomvv1/Resrve/internal/resources"
	"github.com/gin-gonic/gin"
)

var ErrMakingLeaseId = errors.New("Error: unable to make an id for the new lease")

var leaseMap = make(map[*resources.Resource]*Lease)
var muLeaseMap = sync.Mutex{}

type Lease struct {
	ID           string    `json:"id"`
	Holder       string    `json:"holder"`
	FencingToken int       `json:"fencing_token"`
	TTL          time.Time `json:"ttl"`
	done         chan struct{}
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
		TTL:          time.Now().Add(time.Second * 10),
		done:         make(chan struct{}),
	}, nil
}

func getLease(resource *resources.Resource) (*Lease, bool) {
	muLeaseMap.Lock()
	lease, ok := leaseMap[resource]
	muLeaseMap.Unlock()

	return lease, ok
}

func CreateAndReturnLease(c *gin.Context, resource *resources.Resource, fencingToken int) *Lease {
	resultLease, err := NewLease(c.ClientIP(), fencingToken)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return nil
	}

	resource.State = resources.StateTaken

	muLeaseMap.Lock()
	leaseMap[resource] = resultLease
	muLeaseMap.Unlock()

	c.JSON(http.StatusOK, gin.H{"result": resultLease})
	return resultLease
}

func AcquireLease(c *gin.Context) {
	now := c.GetTime("time")
	resourceAny, _ := c.Get("resource")
	resource := resourceAny.(*resources.Resource)

	lease, ok := getLease(resource)
	if !ok {
		resultLease := CreateAndReturnLease(c, resource, 0)
		go cleanState(resultLease, resource)
		return
	}

	if now.After(lease.TTL) {
		resultLease := CreateAndReturnLease(c, resource, lease.FencingToken+1)
		go cleanState(resultLease, resource)
		return
	} else {
		c.JSON(http.StatusConflict, gin.H{"error": "Error: the resource is unavailable at the moment"})
		return
	}
}

func RenewLease(c *gin.Context) {
	now := c.GetTime("time")
	resourceAny, _ := c.Get("resource")
	resource := resourceAny.(*resources.Resource)

	lease, ok := getLease(resource)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Error: there is no lease made on this resource"})
		return
	}

	if lease.Holder == c.ClientIP() && !now.After(lease.TTL) {
		lease.FencingToken++
		lease.TTL = time.Now().UTC().Add(10 * time.Second)
		c.JSON(http.StatusOK, gin.H{"result": lease})
		return
	}

	c.JSON(http.StatusForbidden, gin.H{"error": "Error: you don't have the permission to renew a lease you don't own"})
}

func ReleaseLease(c *gin.Context) {
	now := c.GetTime("time")
	resourceAny, _ := c.Get("resource")
	resource := resourceAny.(*resources.Resource)

	lease, ok := getLease(resource)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Error: there is no lease made on this resource"})
		return
	}

	if lease.Holder == c.ClientIP() && !now.After(lease.TTL) {
		resource.State = resources.StateFree
		close(lease.done)
		c.JSON(http.StatusOK, gin.H{"result": "You have successfully released the lease"})
		return
	}

	c.JSON(http.StatusForbidden, gin.H{"error": "Error: you don't have the permission to release a lease you don't own"})
}

func cleanState(lease *Lease, resource *resources.Resource) {
	for {
		select {
		case <-lease.done:
			return

		default:
			time.Sleep(time.Until(lease.TTL))
			if time.Now().UTC().After(lease.TTL) {
				resource.State = resources.StateFree
				return
			}
		}
	}
}
