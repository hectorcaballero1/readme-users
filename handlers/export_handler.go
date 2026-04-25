package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ms1-users/models"
)

type ExportHandler struct {
	DB *gorm.DB
}

func NewExportHandler(db *gorm.DB) *ExportHandler {
	return &ExportHandler{DB: db}
}

// ExportUsers godoc
// @Summary      Export all users
// @Tags         Export
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {array}   models.UserResponse
// @Failure      500  {object}  map[string]string
// @Router       /export/users [get]
func (h *ExportHandler) ExportUsers(c *gin.Context) {
	var users []models.User
	if err := h.DB.Preload("Zone").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}

	response := make([]models.UserResponse, len(users))
	for i, u := range users {
		response[i] = models.UserResponse{
			ID:        u.ID,
			Name:      u.Name,
			Email:     u.Email,
			ZoneID:    u.ZoneID,
			Zone:      u.Zone,
			PhotoURL:  u.PhotoURL,
			CreatedAt: u.CreatedAt,
		}
	}
	c.JSON(http.StatusOK, response)
}

// ExportZones godoc
// @Summary      Export all zones
// @Tags         Export
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {array}   models.Zone
// @Failure      500  {object}  map[string]string
// @Router       /export/zones [get]
func (h *ExportHandler) ExportZones(c *gin.Context) {
	var zones []models.Zone
	if err := h.DB.Find(&zones).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve zones"})
		return
	}
	c.JSON(http.StatusOK, zones)
}
