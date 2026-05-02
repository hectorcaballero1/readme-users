package routes

import (
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	"ms1-users/handlers"
	"ms1-users/middlewares"
	"ms1-users/storage"
)

func Setup(r *gin.Engine, db *gorm.DB, s3 *storage.S3Service) {
	allowedOrigin := os.Getenv("ALLOWED_ORIGINS")
	var corsConfig cors.Config
	if allowedOrigin == "" {
		corsConfig = cors.DefaultConfig()
		corsConfig.AllowAllOrigins = true
	} else {
		corsConfig = cors.DefaultConfig()
		corsConfig.AllowOrigins = []string{allowedOrigin}
	}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Authorization", "Content-Type", "X-Api-Key"}
	r.Use(cors.New(corsConfig))

	userHandler := handlers.NewUserHandler(db, s3)
	zoneHandler := handlers.NewZoneHandler(db)
	exportHandler := handlers.NewExportHandler(db)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Swagger UI at /swagger/index.html
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api")
	{
		// Public endpoints
		api.POST("/users", userHandler.Register)
		api.POST("/users/login", userHandler.Login)

		// Protected endpoints (require valid JWT)
		auth := api.Group("")
		auth.Use(middlewares.AuthMiddleware())
		{
			auth.GET("/users/:id", userHandler.GetProfile)
			auth.PUT("/users/:id", userHandler.UpdateProfile)
			auth.POST("/users/:id/photo", userHandler.UploadPhoto)
			auth.GET("/zones", zoneHandler.ListZones)
		}

		// Export endpoints (require API key)
		export := api.Group("/export")
		export.Use(middlewares.ApiKeyMiddleware())
		{
			export.GET("/users", exportHandler.ExportUsers)
			export.GET("/zones", exportHandler.ExportZones)
		}
	}
}
