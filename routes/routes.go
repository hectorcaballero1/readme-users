package routes

import (
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
	r.Use(cors.Default())

	userHandler := handlers.NewUserHandler(db, s3)
	zoneHandler := handlers.NewZoneHandler(db)
	exportHandler := handlers.NewExportHandler(db)

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
