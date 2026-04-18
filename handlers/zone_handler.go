package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ms1-users/models"
)

type ZoneHandler struct {
	DB *gorm.DB
}

func NewZoneHandler(db *gorm.DB) *ZoneHandler {
	return &ZoneHandler{DB: db}
}

type ZoneListResponse struct {
	Data    []models.Zone `json:"data"`
	Message string        `json:"message"`
}

// ListZones godoc
// @Summary      List zones
// @Description  Returns all available delivery/pickup zones for dropdown population in the frontend.
// @Tags         zones
// @Produce      json
// @Success      200  {object}  ZoneListResponse       "List of zones"
// @Failure      401  {object}  models.ErrorResponse   "Unauthorized"
// @Security     BearerAuth
// @Router       /zones [get]
func (h *ZoneHandler) ListZones(c *gin.Context) {
	var zones []models.Zone
	h.DB.Find(&zones)

	c.JSON(http.StatusOK, gin.H{
		"data":    zones,
		"message": "Zones retrieved successfully",
	})
}
