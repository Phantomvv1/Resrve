package middleware

import (
	"net/http"
	"time"

	"github.com/Phantomvv1/Resrve/internal/resources"
	"github.com/gin-gonic/gin"
)

func ResourceGetterMiddleware(c *gin.Context) {
	now := time.Now().UTC()
	id := c.Param("id")

	resource, err := resources.GetResourceById(id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Set("time", now)
	c.Set("resource", resource)

	c.Next()
}
