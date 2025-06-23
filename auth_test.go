package main

import (
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Setup a test router and in-memory DB
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.Default()
	r.LoadHTMLGlob("tempelates/*") // Adjust path to test HTML templates

	// Sessions
	store := cookie.NewStore([]byte("test-secret"))
	r.Use(sessions.Sessions("test-session", store))

	// In-memory SQLite DB for isolation
	var err error
	DB, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to test database")
	}
	DB.AutoMigrate(&User{})

	// Routes to test
	r.POST("/register", RegisterUser)
	r.POST("/login", LoginUser)

	return r
}

func TestRegisterUser(t *testing.T) {
	router := setupTestRouter()

	form := url.Values{}
	form.Add("email", "testuser@example.com")
	form.Add("password", "password123")

	req := httptest.NewRequest("POST", "/register", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/login", w.Header().Get("Location"))
}

func TestLoginUser_Success(t *testing.T) {
	router := setupTestRouter()

	// Insert test user manually
	password := "securepassword"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	DB.Create(&User{Email: "loginuser@example.com", Password: string(hashed)})

	form := url.Values{}
	form.Add("email", "loginuser@example.com")
	form.Add("password", "securepassword")

	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/upload", w.Header().Get("Location"))
}

func TestLoginUser_InvalidPassword(t *testing.T) {
	router := setupTestRouter()

	password := "securepassword"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	DB.Create(&User{Email: "wrongpass@example.com", Password: string(hashed)})

	form := url.Values{}
	form.Add("email", "wrongpass@example.com")
	form.Add("password", "wrongpassword")

	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid password")
}

func TestLoginUser_UserNotFound(t *testing.T) {
	router := setupTestRouter()

	form := url.Values{}
	form.Add("email", "notfound@example.com")
	form.Add("password", "whatever")

	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "User not found")
}
