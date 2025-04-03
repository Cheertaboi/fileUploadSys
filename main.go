package main

import (
	"fmt"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() { // 👈 Make sure this function exists!
	InitDB()

	r := gin.Default()
	r.LoadHTMLGlob("D:/go/loginpage/tempelates/*")

	// Session middleware
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("session", store))

	// Routes
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "home.html", nil) // Make sure "home.html" exists
	})
	r.GET("/register", ShowRegisterPage)
	r.POST("/register", RegisterUser)
	r.GET("/login", ShowLoginPage)
	r.POST("/login", LoginUser)
	r.GET("/home", ShowHomePage)
	r.GET("/logout", LogoutUser)
	r.GET("/upload", ShowUploadPage)
	r.POST("/upload", UploadFile)
	r.GET("/files", ListFiles)
	r.GET("/download/:id", DownloadFile)
	r.GET("/delete/:id", DeleteFile)
	r.GET("/share/:id", ShareFile)
	r.GET("/:tinyURL", RedirectToFile)

	// Cheacking for all the routes
	for _, route := range r.Routes() {
		fmt.Printf("Path: %s | Method: %s | Handlers: %d\n", route.Path, route.Method, len(route.Handler))

	}

	// Start server

	r.Run(":8080")
}
