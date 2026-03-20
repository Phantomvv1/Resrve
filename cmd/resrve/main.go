package main

import (
	"net/http"

	"github.com/Phantomvv1/Resrve/internal/leases"
	"github.com/Phantomvv1/Resrve/internal/middleware"
	"github.com/Phantomvv1/Resrve/internal/resources"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) { c.JSON(http.StatusOK, nil) })

	resource := r.Group("/resources")
	resource.GET("/:id", middleware.ResourceGetterMiddleware, resources.GetResource)
	resource.GET("", resources.GetAllResources)

	lease := r.Group("/leases")
	lease.Use(middleware.ResourceGetterMiddleware)

	lease.POST("/:id", leases.AcquireLease)
	lease.DELETE("/:id", leases.ReleaseLease)
	lease.POST("/:id/renew", leases.RenewLease)

	r.Run(":42069")
}
