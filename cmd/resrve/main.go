package main

import (
	"net/http"

	"github.com/Phantomvv1/Resrve/internal/leases"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) { c.JSON(http.StatusOK, nil) })

	lease := r.Group("/leases")
	lease.GET("/:id", leases.GetLease)
	lease.GET("", leases.GetAllLeases)
	lease.POST("/:id", leases.AcquireLease)
	lease.DELETE("/:id", leases.ReleaseLease)
	lease.POST("/:id/renew", leases.RenewLease)

	r.Run(":42069")
}
