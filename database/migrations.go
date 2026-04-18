package database

import (
	"log"

	"gorm.io/gorm"

	"ms1-users/models"
)

var seedZones = []string{
	"Miraflores", "San Isidro", "Barranco", "San Miguel", "Surco",
	"La Molina", "Jesús María", "Magdalena", "Pueblo Libre", "Lince",
	"San Borja", "Breña", "Chorrillos", "Callao", "Surquillo",
	"Santa Anita", "Ate", "San Juan de Miraflores", "Villa El Salvador",
	"Los Olivos", "San Martín de Porres", "Independencia", "Comas",
	"El Agustino", "La Victoria", "Rímac", "Otros",
}

func RunMigrations(db *gorm.DB) {
	if err := db.AutoMigrate(&models.Zone{}, &models.User{}); err != nil {
		log.Fatal("Migration failed: ", err)
	}
	log.Println("Migrations applied")

	var count int64
	db.Model(&models.Zone{}).Count(&count)
	if count == 0 {
		for _, name := range seedZones {
			db.Create(&models.Zone{Name: name})
		}
		log.Printf("Seeded %d zones", len(seedZones))
	}
}
