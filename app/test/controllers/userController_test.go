package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/gin-gonic/gin"
	"github.com/sd0ric4/book-reader-backend/app/config"
	"github.com/sd0ric4/book-reader-backend/app/controllers"
	"github.com/sd0ric4/book-reader-backend/app/database"
	"github.com/sd0ric4/book-reader-backend/app/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() {
	var err error
	database.MySQLDB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	database.MySQLDB.AutoMigrate(&models.User{})
}

func TestRegister(t *testing.T) {
	setupTestDB()

	router := gin.Default()
	router.POST("/register", controllers.Register)

	user := models.User{
		Email:    "test@example.com",
		Password: "password123",
	}

	jsonValue, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status code %d but got %d", http.StatusOK, w.Code)
	}

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["message"] != "User registered successfully" {
		t.Fatalf("Expected message 'User registered successfully' but got '%s'", response["message"])
	}

	var dbUser models.User
	if err := database.MySQLDB.Where("email = ?", user.Email).First(&dbUser).Error; err != nil {
		t.Fatalf("Expected user to be created in the database, but got error: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte("password123")); err != nil {
		t.Fatalf("Expected password to be hashed correctly, but got error: %v", err)
	}
}

func TestLogin(t *testing.T) {
	setupTestDB()
	config.LoadConfig("../../../config/config.yaml")
	router := gin.Default()
	router.POST("/login", controllers.Login)

	user := models.User{
		Email:    "test@example.com",
		Password: "password123",
	}
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(hashedPassword)

	database.MySQLDB.Create(&user)

	user.Password = "password123"
	jsonValue, _ := json.Marshal(user)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status code %d but got %d", http.StatusOK, w.Code)
	}

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["message"] != "Login successful" {
		t.Fatalf("Expected message 'Login successful' but got '%s'", response["message"])
	}

}
func TestRealRegister(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	config.Config = &config.ConfigStruct{
		MySQL: config.MySQLConfig{
			User:     "root",
			Password: "root",
			Host:     "192.168.0.118",
			Port:     3306,
			DBName:   "bookdb",
		},
	}
	database.InitMySQL()

	router := gin.Default()
	router.POST("/register", controllers.Register)

	user := models.User{
		Email:    faker.Email(),
		Password: "password123",
		Username: faker.Username(),
	}

	userJSON, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(userJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status code %d but got %d", http.StatusOK, w.Code)
	}

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["message"] != "User registered successfully" {
		t.Fatalf("Expected message 'User registered successfully' but got '%s'", response["message"])
	}

	var dbUser models.User
	if err := database.MySQLDB.Where("email = ?", user.Email).First(&dbUser).Error; err != nil {
		t.Fatalf("Expected user to be created in the database, but got error: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte("password123")); err != nil {
		t.Fatalf("Expected password to be hashed correctly, but got error: %v", err)
	}
}

func TestRealLogin(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	config.Config = &config.ConfigStruct{
		MySQL: config.MySQLConfig{
			User:     "root",
			Password: "root",
			Host:     "192.168.0.118",
			Port:     3306,
			DBName:   "bookdb",
		},
	}
	database.InitMySQL()

	router := gin.Default()
	router.POST("/login", controllers.Login)

	user := models.User{
		Email:    "mcwcrlO@gYfjTUd.org",
		Password: "password123",
	}

	userJSON, _ := json.Marshal(user)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(userJSON))

	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status code %d but got %d", http.StatusOK, w.Code)
	}

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["message"] != "Login successful" {
		t.Fatalf("Expected message 'Login successful' but got '%s'", response["message"])
	}

}
