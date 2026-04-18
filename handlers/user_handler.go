package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"ms1-users/middlewares"
	"ms1-users/models"
	"ms1-users/storage"
)

type UserHandler struct {
	DB      *gorm.DB
	Storage *storage.S3Service
}

func NewUserHandler(db *gorm.DB, s3 *storage.S3Service) *UserHandler {
	return &UserHandler{DB: db, Storage: s3}
}

// ── Request structs ──────────────────────────────────────────────────────────

type RegisterRequest struct {
	Name     string `json:"name"     binding:"required"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	ZoneID   *uint  `json:"zone_id"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func toUserResponse(u models.User) models.UserResponse {
	return models.UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		ZoneID:    u.ZoneID,
		Zone:      u.Zone,
		PhotoURL:  u.PhotoURL,
		CreatedAt: u.CreatedAt,
	}
}

// ── Handlers ─────────────────────────────────────────────────────────────────

// Register godoc
// @Summary      Register a new user
// @Description  Creates a user account. zone_id and photo_url are optional.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        input  body      RegisterRequest        true  "Registration data"
// @Success      201    {object}  map[string]interface{} "User created"
// @Failure      400    {object}  models.ErrorResponse   "Validation error"
// @Failure      409    {object}  models.ErrorResponse   "Email already registered"
// @Failure      500    {object}  models.ErrorResponse   "Internal error"
// @Router       /users [post]
func (h *UserHandler) Register(c *gin.Context) {
	var input RegisterRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.User
	if h.DB.Where("email = ?", input.Email).First(&existing).Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing password"})
		return
	}

	user := models.User{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: string(hash),
		ZoneID:       input.ZoneID,
	}

	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating user"})
		return
	}

	// Preload zone association if set
	if user.ZoneID != nil {
		h.DB.Preload("Zone").First(&user, user.ID)
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":    toUserResponse(user),
		"message": "User registered successfully",
	})
}

// Login godoc
// @Summary      Login
// @Description  Authenticates a user and returns a JWT valid for 24 h.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        input  body      LoginRequest           true  "Credentials"
// @Success      200    {object}  map[string]interface{} "Token + user"
// @Failure      400    {object}  models.ErrorResponse   "Validation error"
// @Failure      401    {object}  map[string]interface{} "Invalid credentials"
// @Failure      500    {object}  models.ErrorResponse   "Internal error"
// @Router       /users/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var input LoginRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if h.DB.Where("email = ?", input.Email).First(&user).Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	claims := middlewares.Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}

	// Preload zone for response
	if user.ZoneID != nil {
		h.DB.Preload("Zone").First(&user, user.ID)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"token": tokenStr,
			"user":  toUserResponse(user),
		},
		"message": "Login successful",
	})
}

// GetProfile godoc
// @Summary      Get user profile
// @Description  Returns a user's public profile. Never includes password_hash. Called by other microservices for user validation.
// @Tags         users
// @Produce      json
// @Param        id   path      int                    true  "User ID"
// @Success      200  {object}  map[string]interface{} "User profile"
// @Failure      400  {object}  models.ErrorResponse   "Invalid ID"
// @Failure      401  {object}  models.ErrorResponse   "Unauthorized"
// @Failure      404  {object}  models.ErrorResponse   "User not found"
// @Security     BearerAuth
// @Router       /users/{id} [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user models.User
	if h.DB.Preload("Zone").First(&user, id).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    toUserResponse(user),
		"message": "User found",
	})
}

// UpdateProfile godoc
// @Summary      Update user profile
// @Description  Updates name, zone_id, and/or photo_url. Only the account owner can call this endpoint. Pass null to clear a nullable field.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id     path      int                    true  "User ID"
// @Param        input  body      object                 true  "Fields to update (name, zone_id, photo_url)"
// @Success      200    {object}  map[string]interface{} "Updated profile"
// @Failure      400    {object}  models.ErrorResponse   "Validation error"
// @Failure      401    {object}  models.ErrorResponse   "Unauthorized"
// @Failure      403    {object}  models.ErrorResponse   "Forbidden"
// @Failure      404    {object}  models.ErrorResponse   "User not found"
// @Failure      500    {object}  models.ErrorResponse   "Internal error"
// @Security     BearerAuth
// @Router       /users/{id} [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Ownership check: JWT user_id must match path :id
	tokenUserID, _ := c.Get("user_id")
	if tokenUserID.(uint) != uint(id) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only edit your own profile"})
		return
	}

	var user models.User
	if h.DB.First(&user, id).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Parse raw JSON so we can distinguish "field absent" from "field = null"
	var rawBody map[string]json.RawMessage
	if err := c.ShouldBindJSON(&rawBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}

	if v, ok := rawBody["name"]; ok {
		var name string
		if err := json.Unmarshal(v, &name); err == nil && name != "" {
			updates["name"] = name
		}
	}

	if v, ok := rawBody["zone_id"]; ok {
		if string(v) == "null" {
			// Explicit null clears the zone association
			updates["zone_id"] = nil
		} else {
			var zoneID uint
			if err := json.Unmarshal(v, &zoneID); err == nil && zoneID > 0 {
				updates["zone_id"] = zoneID
			}
		}
	}

	if v, ok := rawBody["photo_url"]; ok {
		if string(v) == "null" {
			if h.Storage != nil && user.PhotoURL != nil {
				h.Storage.DeleteProfilePhoto(*user.PhotoURL)
			}
			updates["photo_url"] = nil
		} else {
			var photoURL string
			if err := json.Unmarshal(v, &photoURL); err == nil {
				updates["photo_url"] = photoURL
			}
		}
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields to update"})
		return
	}

	if err := h.DB.Model(&user).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating profile"})
		return
	}

	h.DB.Preload("Zone").First(&user, id)

	c.JSON(http.StatusOK, gin.H{
		"data":    toUserResponse(user),
		"message": "Profile updated successfully",
	})
}

// UploadPhoto godoc
// @Summary      Upload profile photo
// @Description  Uploads an image to S3 and stores the URL as the user's photo. Replaces the previous photo if one exists.
// @Tags         users
// @Accept       multipart/form-data
// @Produce      json
// @Param        id     path      int  true  "User ID"
// @Param        photo  formData  file true  "Photo file"
// @Success      200    {object}  map[string]interface{} "Updated profile"
// @Failure      400    {object}  models.ErrorResponse   "Missing file"
// @Failure      401    {object}  models.ErrorResponse   "Unauthorized"
// @Failure      403    {object}  models.ErrorResponse   "Forbidden"
// @Failure      404    {object}  models.ErrorResponse   "User not found"
// @Failure      500    {object}  models.ErrorResponse   "Upload error"
// @Failure      503    {object}  models.ErrorResponse   "Storage not configured"
// @Security     BearerAuth
// @Router       /users/{id}/photo [post]
func (h *UserHandler) UploadPhoto(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	tokenUserID, _ := c.Get("user_id")
	if tokenUserID.(uint) != uint(id) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only edit your own profile"})
		return
	}

	if h.Storage == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Storage service not configured"})
		return
	}

	fileHeader, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Field 'photo' is required"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading file"})
		return
	}
	defer file.Close()

	var user models.User
	if h.DB.First(&user, id).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	newURL, err := h.Storage.UploadProfilePhoto(file, fileHeader, uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error uploading photo"})
		return
	}

	// Delete previous photo from S3 (best-effort)
	if user.PhotoURL != nil {
		h.Storage.DeleteProfilePhoto(*user.PhotoURL)
	}

	if err := h.DB.Model(&user).Update("photo_url", newURL).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving photo URL"})
		return
	}

	h.DB.Preload("Zone").First(&user, id)

	c.JSON(http.StatusOK, gin.H{
		"data":    toUserResponse(user),
		"message": "Photo uploaded successfully",
	})
}
