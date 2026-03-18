package leases

import (
	"net/http"

	"github.com/Phantomvv1/Resrve/internal/resources"
	"github.com/gin-gonic/gin"
)

var fencingTokenMap = make(map[resources.Resource]int)

func AcquireLease(c *gin.Context) {

}

func RenewLease(c *gin.Context) {

}

func ReleaseLease(c *gin.Context) {

}

func GetLease(c *gin.Context) {
	id := c.Param("id")

	for _, resource := range resources.AvailableResources {
		if resource.ID == id {
			c.JSON(http.StatusOK, gin.H{"resource": resource})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Error: there is no resource with that id"})
}

func GetAllLeases(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"leases": resources.AvailableResources})
}
