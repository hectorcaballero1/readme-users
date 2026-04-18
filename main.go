// @title           MS1 Users API
// @version         1.0
// @description     User management microservice for the ReadMe book marketplace platform.
// @host            localhost:8001
// @BasePath        /api

// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
// @description                Enter: Bearer <token>

package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"ms1-users/config"
	"ms1-users/database"
	_ "ms1-users/docs"
	"ms1-users/routes"
	"ms1-users/storage"
)

func main() {
	cfg := config.Load()

	db := database.Connect(cfg.DSN())
	database.RunMigrations(db)

	var s3 *storage.S3Service
	if cfg.AWSAccessKey != "" && cfg.AWSBucket != "" {
		var err error
		s3, err = storage.NewS3Service(cfg.AWSAccessKey, cfg.AWSSecretKey, cfg.AWSRegion, cfg.AWSBucket)
		if err != nil {
			log.Printf("WARNING: S3 storage not available: %v", err)
		}
	} else {
		log.Println("WARNING: AWS credentials not set — photo upload disabled")
	}

	r := gin.Default()
	routes.Setup(r, db, s3)

	log.Printf("Server running on :%s  — Swagger: http://localhost:%s/swagger/index.html", cfg.Port, cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Server failed to start: ", err)
	}
}
