package leases

import (
	"log"
	"net/http"
	"time"

	"github.com/Phantomvv1/Resrve/internal/resources"
	"github.com/gin-gonic/gin"
)

func CreateAndReturnLease(c *gin.Context, resource *resources.Resource, fencingToken int) {
	lease, err := resources.NewLease(c.ClientIP(), fencingToken)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	resource.State = resources.StateTaken
	resource.UpdateLease(lease)

	c.JSON(http.StatusOK, gin.H{"result": lease})
}

func AcquireLease(c *gin.Context) {
	now := c.GetTime("time")
	resourceAny, _ := c.Get("resource")
	resource := resourceAny.(*resources.Resource)

	resource.Lock()
	defer resource.Unlock()

	lease := resource.Lease()
	if lease == nil {
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
	now := c.GetTime("time")
	resourceAny, _ := c.Get("resource")
	resource := resourceAny.(*resources.Resource)
	fencingToken := c.GetInt("fencingToken")

	resource.Lock()
	defer resource.Unlock()

	lease := resource.Lease()
	if lease == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Error: there is no lease made on this resource"})
		return
	}

	if lease.FencingToken != fencingToken {
		c.JSON(http.StatusForbidden, gin.H{"error": "Error: you don't have the permission to renew a lease you don't own. Fencing token is different"})
		return
	}

	if lease.Holder == c.ClientIP() && !now.After(lease.TTL) {
		lease.FencingToken++
		lease.TTL = now.Add(10 * time.Second)

		c.JSON(http.StatusOK, gin.H{"result": lease})
		return
	} else if now.After(lease.TTL) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Error: you can't renew a lease that has already expired"})
		return
	}

	c.JSON(http.StatusForbidden, gin.H{"error": "Error: you don't have the permission to renew a lease you don't own"})
}

func ReleaseLease(c *gin.Context) {
	now := c.GetTime("time")
	resourceAny, _ := c.Get("resource")
	resource := resourceAny.(*resources.Resource)
	fencingToken := c.GetInt("fencingToken")

	resource.Lock()
	defer resource.Unlock()

	lease := resource.Lease()
	if lease == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Error: there is no lease made on this resource"})
		return
	}

	if lease.FencingToken != fencingToken {
		c.JSON(http.StatusForbidden, gin.H{"error": "Error: you don't have the permission to release a lease you don't own. Fencing token is different"})
		return
	} else if now.After(lease.TTL) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Error: you can't release a lease that has already expired"})
		return
	}

	if lease.Holder == c.ClientIP() && !now.After(lease.TTL) {
		resource.UpdateLease(nil)
		resource.State = resources.StateFree
		c.JSON(http.StatusOK, gin.H{"result": "You have successfully released the lease"})
		return
	}

	c.JSON(http.StatusForbidden, gin.H{"error": "Error: you don't have the permission to release a lease you don't own"})
}
