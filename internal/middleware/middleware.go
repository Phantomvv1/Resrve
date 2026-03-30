package middleware

import (
	"encoding/json"
	"log"
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

func JSONParserMiddleware(c *gin.Context) {
	var information map[string]any
	err := json.NewDecoder(c.Request.Body).Decode(&information)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Error: unable to decode the json body of the request"})
		return
	}

	for k, v := range information {
		c.Set(k, v)
	}
}
