package v1

import (
	"database/sql"
	"net/http"

	"ppl-m1-be/config"

	"github.com/gin-gonic/gin"

	apidocs "ppl-m1-be/docs"
)

func NewRouter(db *sql.DB, cfg config.Config) *gin.Engine {
	router := gin.Default()
	
	router.GET("/openapi.yaml", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml; charset=utf-8", apidocs.OpenAPI)
	})
	
	router.GET("/api/m1/v1/docs", docsHandler)
	
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	router.GET("/ready", func(c *gin.Context) {
		if err := db.PingContext(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
				"error":  "database connection failed",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   "ready",
			"database": "connected",
		})
	})

	api := router.Group("/api/m1/v1")
	mountAuth(api, db, cfg)
	mountUser(api, db)

	return router
}