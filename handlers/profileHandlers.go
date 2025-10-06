package handlers

import (
	"day/database"
	"day/models"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetProfile(c *gin.Context) {
	id := c.Param("id")

	var user models.User

	if err := database.DB.
		Preload("Profile").
		Preload("Books.Creator").
		First(&user, id).Error; err != nil {

		c.JSON(http.StatusNotFound, gin.H{"error": "профиль не найден"})
		return
	}

	type CreatorResponse struct {
		ID       uint   `json:"id"`
		UserName string `json:"username"`
	}

	type BookResponse struct {
		ID      uint            `json:"id"`
		Name    string          `json:"name"`
		Author  string          `json:"author"`
		Image   string          `json:"image"`
		Creator CreatorResponse `json:"creator"`
	}

	var books []BookResponse
	for _, b := range user.Books {
		books = append(books, BookResponse{
			ID:     b.ID,
			Name:   b.Name,
			Author: b.Author,
			Image:  b.Image,
			Creator: CreatorResponse{
				ID:       b.Creator.ID,
				UserName: b.Creator.UserName,
			},
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"username": user.UserName,
		"profile": gin.H{
			"avatar": user.Profile.Avatar,
			"bio":    user.Profile.Bio,
		},
		"books": books,
	})
}

func UpdateProfile(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "пользователь не авторизован"})
		return
	}

	var userID uint
	switch v := userIDInterface.(type) {
	case float64:
		userID = uint(v)
	case uint:
		userID = v
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "неверный формат user_id"})
		return
	}

	bio := c.PostForm("bio")

	file, err := c.FormFile("avatar")
	var filepathStr string
	if err == nil {
		if err := os.MkdirAll("uploads/avatars", os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось создать папку для аватарок"})
			return
		}

		ext := filepath.Ext(file.Filename)
		uniqueName := uuid.New().String() + ext
		filepathStr = "uploads/avatars/" + uniqueName

		if err := c.SaveUploadedFile(file, filepathStr); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "не удалось сохранить аватар"})
			return
		}
	}

	var profile models.Profile
	if err := database.DB.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		profile = models.Profile{
			UserID: userID,
			Bio:    bio,
			Avatar: filepathStr,
		}
		database.DB.Create(&profile)
	} else {
		if filepathStr != "" {
			profile.Avatar = filepathStr
		}
		profile.Bio = bio
		database.DB.Save(&profile)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "профиль обновлён",
		"profile": gin.H{
			"avatar": profile.Avatar,
			"bio":    profile.Bio,
		},
	})
}
