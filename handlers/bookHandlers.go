package handlers

import (
	"day/database"
	"day/models"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserResponse struct {
	ID       uint   `json:"id"`
	UserName string `json:"username"`
}

type BookResponse struct {
	ID      uint         `json:"id"`
	Name    string       `json:"name"`
	Author  string       `json:"author"`
	Image   string       `json:"image"`
	Creator UserResponse `json:"creator"`
}

func CreateBook(c *gin.Context) {
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

	name := c.PostForm("name")
	author := c.PostForm("author")
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := os.MkdirAll("uploads/books", os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось создать папку для книг"})
		return
	}

	ext := filepath.Ext(file.Filename)
	uniqueName := uuid.New().String() + ext
	filepathStr := "uploads/books/" + uniqueName

	if err := c.SaveUploadedFile(file, filepathStr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	book := models.Book{
		Name:   name,
		Author: author,
		Image:  filepathStr,
		UserID: userID,
	}

	if err := database.DB.Create(&book).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	database.DB.Preload("Creator").First(&book, book.ID)

	resp := BookResponse{
		ID:     book.ID,
		Name:   book.Name,
		Author: book.Author,
		Image:  book.Image,
		Creator: UserResponse{
			ID:       book.Creator.ID,
			UserName: book.Creator.UserName,
		},
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "книга успешно создана",
		"book":    resp,
	})
}

func GetBooks(c *gin.Context) {
	var books []models.Book

	if err := database.DB.Find(&books).Error; err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, books)
}

func GetBookByID(c *gin.Context) {
	var book models.Book
	idstr := c.Param("id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	if err := database.DB.Where("id = ?", id).First(&book).Error; err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, book)
}

func DeleteBook(c *gin.Context) {

	var book models.Book
	idstr := c.Param("id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	if err := database.DB.Where("id = ?", id).First(&book).Error; err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	if err := database.DB.Delete(&book).Error; err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Book deleted succesfully"})
}

func PaginateBooks(c *gin.Context) {
	var books []models.Book

	countStr := c.Param("count")
	count, _ := strconv.Atoi(countStr)

	if err := database.DB.Preload("Creator").Limit(count).Find(&books).Error; err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	var result []BookResponse
	for _, b := range books {
		result = append(result, BookResponse{
			ID:     b.ID,
			Name:   b.Name,
			Author: b.Author,
			Image:  b.Image,
			Creator: UserResponse{
				ID:       b.Creator.ID,
				UserName: b.Creator.UserName,
			},
		})
	}

	c.JSON(http.StatusOK, result)
}

func UpdateBook(c *gin.Context) {
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

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID книги"})
		return
	}

	var book models.Book
	if err := database.DB.Preload("Creator").First(&book, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "книга не найдена"})
		return
	}

	if book.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "вы не можете редактировать чужую книгу"})
		return
	}

	name := c.PostForm("name")
	author := c.PostForm("author")
	file, err := c.FormFile("image")

	if name != "" {
		book.Name = name
	}
	if author != "" {
		book.Author = author
	}

	if err == nil {
		if err := os.MkdirAll("uploads/books", os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось создать папку для книг"})
			return
		}

		ext := filepath.Ext(file.Filename)
		uniqueName := uuid.New().String() + ext
		filepathStr := "uploads/books/" + uniqueName

		if err := c.SaveUploadedFile(file, filepathStr); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		book.Image = filepathStr
	}

	if err := database.DB.Save(&book).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := BookResponse{
		ID:     book.ID,
		Name:   book.Name,
		Author: book.Author,
		Image:  book.Image,
		Creator: UserResponse{
			ID:       book.Creator.ID,
			UserName: book.Creator.UserName,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "книга успешно обновлена",
		"book":    resp,
	})
}
