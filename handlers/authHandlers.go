package handlers

import (
	"day/database"
	"day/models"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var Secret = os.Getenv("SECRET")

type RegisterData struct {
	UserName  string `json:"username" validate:"required,min=3,max=30"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	Password2 string `json:"password2" validate:"required,eqfield=Password"`
}

type LoginData struct {
	UserName string `json:"username" validate:"required,min=3,max=30"`
	Password string `json:"password" validate:"required,min=8"`
}

var validate = validator.New()

func Register(c *gin.Context) {
	var user RegisterData
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validate.Struct(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.User

	if err := database.DB.Where("email = ?", user.Email).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "пользователь с таким email уже существует"})
		return
	} else if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка при проверке email"})
		return
	}

	if err := database.DB.Where("user_name = ?", user.UserName).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "пользователь с таким username уже существует"})
		return
	} else if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка при проверке username"})
		return
	}

	hashPass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка при генерации пароля"})
		return
	}

	newUser := models.User{
		UserName: user.UserName,
		Email:    user.Email,
		Password: string(hashPass),
	}

	if err := database.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось создать пользователя"})
		return
	}

	token, err := GenerateToken(newUser.ID, newUser.UserName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось сгенерировать токен"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "вы успешно зарегистрировались!",
		"token":   token,
	})
}

func Login(c *gin.Context) {
	var input LoginData

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный запрос данных"})
		return
	}

	if err := validate.Struct(&input); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	var existing models.User

	if err := database.DB.Where("user_name = ?", input.UserName).First(&existing).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "вы ввели неверный логин"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(existing.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "неверный пароль"})
		return
	}

	token, err := GenerateToken(existing.ID, existing.UserName)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось сгенерировать токен"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "вы успешно авторизировались",
		"token":   token,
	})
}

func GenerateToken(UserID uint, Username string) (string, error) {
	if Secret == "" {
		Secret = "dpifuhgdifugdifungidfngdifgdo"
	}

	claims := jwt.MapClaims{
		"user_id":  UserID,
		"username": Username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(Secret))
}
