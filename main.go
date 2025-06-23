package main

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	InitDB()

	r := gin.Default()
	r.Static("/static", "./static")
	r.LoadHTMLGlob("C:/Users/HarshVardhana/Music/login/fileUploadSys/tempelates/*")

	// Session middleware
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("session", store))

	// Public routes
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.html", nil)
	})

	r.GET("/register", ShowRegisterPage)
	r.POST("/register", RegisterUser)
	r.GET("/login", ShowLoginPage)
	r.POST("/login", LoginUser)
	r.GET("/logout", LogoutUser)

	// Forgot Password (Microservice Integration)
	r.GET("/forgot-password", ShowForgotPasswordPage)
	r.POST("/forgot-password", HandleForgotPassword)
	r.GET("/reset", ShowResetPasswordPage)
	r.POST("/reset", HandleResetPassword)

	// Protected routes
	auth := r.Group("/")
	auth.Use(AuthRequired())
	{
		auth.GET("/home", ShowHomePage)
		auth.GET("/upload", ShowUploadPage)
		auth.POST("/upload", UploadFile)
		auth.GET("/files", ListFiles)
		auth.GET("/download/:uniqueID", DownloadFile)
		auth.GET("/delete/:id", DeleteFile)
		auth.GET("/share/:id", ShareFile)
	}

	// Public tiny URL route
	r.GET("/:tinyURL", RedirectToFile)

	// Debug all routes
	for _, route := range r.Routes() {
		fmt.Printf("Path: %s | Method: %s\n", route.Path, route.Method)
	}

	r.Run(":8080")

}
