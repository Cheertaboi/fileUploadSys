package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Show registration form
func ShowRegisterPage(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", nil)
	fmt.Println("inside showRegisterPage")
}

// Register new user
func RegisterUser(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := User{Email: email, Password: string(hashedPassword)}
	fmt.Println("inside RegisterUser")
	if err := DB.Create(&user).Error; err != nil {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{"error": "Email already exists"})
		return
	}

	c.Redirect(http.StatusSeeOther, "/login")
}

// Show login form
func ShowLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
	fmt.Println("inside ShowLoginPage")
}

// Authenticate user
func LoginUser(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")
	fmt.Println("inside LoginUser")
	var user User
	if err := DB.Where("email = ?", email).First(&user).Error; err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"error": "User not found"})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"error": "Invalid password"})
		return
	}

	// Set session
	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Save()

	c.Redirect(http.StatusSeeOther, "/upload")
}

// Show home page (after login)
func ShowHomePage(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	if userID == nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	c.HTML(http.StatusOK, "home.html", nil)
}
func ShowUploadPage(c *gin.Context) {
	c.HTML(http.StatusOK, "upload.html", nil)
}
func generateTinyURL() string {
	b := make([]byte, 6) // 6 bytes will give a short URL
	_, err := rand.Read(b)
	if err != nil {
		return "defaultTinyURL" // Fallback in case of error
	}
	return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
}

// UploadFile - Handles file uploads
func UploadFile(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	if userID == nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	// Parse form fields
	title := c.PostForm("title")
	description := c.PostForm("description")

	// Get file from request
	file, err := c.FormFile("file")
	if err != nil {
		c.HTML(http.StatusBadRequest, "upload.html", gin.H{"error": "Failed to upload file"})
		return
	}

	// Generate unique filename
	uniqueID := uuid.New().String()
	fileExt := filepath.Ext(file.Filename)
	newFileName := uniqueID + fileExt
	filePath := "uploads/" + newFileName

	// Save file to disk
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.HTML(http.StatusInternalServerError, "upload.html", gin.H{"error": "Failed to save file"})
		return
	}

	// Ensure unique tiny_url
	var existingFile File
	tinyURL := ""
	attempts := 0
	maxAttempts := 5

	for attempts < maxAttempts {
		tinyURL = generateTinyURL() // Replace this with your tiny URL generation logic

		// Check if tinyURL already exists
		if err := DB.Where("tiny_url = ?", tinyURL).First(&existingFile).Error; err != nil {
			// If no record is found, we can use this tinyURL
			break
		}

		attempts++
	}

	// If we couldn't generate a unique tinyURL, return an error
	if attempts == maxAttempts {
		c.HTML(http.StatusInternalServerError, "upload.html", gin.H{"error": "Failed to generate unique tiny URL"})
		return
	}

	// Save file info to database
	newFile := File{
		UserID:      userID.(uint),
		Title:       title,
		Description: description,
		FilePath:    filePath,
		UniqueID:    uniqueID,
		TinyURL:     tinyURL, // Inserted unique tiny_url here
	}

	if err := DB.Create(&newFile).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "upload.html", gin.H{"error": "Failed to save file metadata"})
		return
	}

	// Redirect to files list
	c.Redirect(http.StatusSeeOther, "/files")
}

// ListFiles - Displays a user's uploaded files
func ListFiles(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	if userID == nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	// Convert userID to uint
	userIDUint, ok := userID.(uint)
	if !ok {
		fmt.Println("Error: Failed to convert userID to uint")
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	var files []File
	err := DB.Where("user_id = ?", userIDUint).Order("created_at DESC").Find(&files).Error
	if err != nil {
		fmt.Println("Error fetching files:", err)
		c.HTML(http.StatusInternalServerError, "files.html", gin.H{"error": "Failed to load files"})
		return
	}

	// Debugging log
	fmt.Println("Files fetched:", files)

	c.HTML(http.StatusOK, "files.html", gin.H{"files": files})
}

// DownloadFile - Allows user to download a file
func DownloadFile(c *gin.Context) {
	uniqueID := c.Param("uniqueID")
	fmt.Println("Attempting to download file with unique ID:", uniqueID)
	if uniqueID == "" {
		fmt.Println("Error: uniqueID is empty")
		c.String(http.StatusBadRequest, "Invalid file ID")
		return
	}
	var file File
	err := DB.Where("unique_id = ?", uniqueID).First(&file).Error
	if err != nil {
		fmt.Println("Error: file not found for unique ID", uniqueID, err)
		c.String(http.StatusNotFound, "File not found")
		return
	}

	// Serve the file
	c.File(file.FilePath)
}

// DeleteFile - Allows user to delete a file
func DeleteFile(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	if userID == nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	fileID := c.Param("id")
	var file File
	if err := DB.Where("unique_id = ? AND user_id = ?", fileID, userID).First(&file).Error; err != nil {
		c.HTML(http.StatusNotFound, "files.html", gin.H{"error": "File not found"})
		return
	}

	// Remove file from disk
	if err := os.Remove(file.FilePath); err != nil {
		fmt.Println("Error deleting file:", err)
	}

	// Delete file from database
	DB.Delete(&file)

	c.Redirect(http.StatusSeeOther, "/files")
}
func RedirectToFile(c *gin.Context) {
	tinyURL := c.Param("tinyURL")

	var file File
	err := DB.Where("tiny_url = ?", tinyURL).First(&file).Error
	if err != nil {
		fmt.Println("Error finding file for tiny URL:", tinyURL, err)
		c.String(http.StatusNotFound, "File not found")
		return
	}

	// Redirect to file download page
	c.Redirect(http.StatusSeeOther, "/download/"+file.UniqueID)
}

// ShareFile - Generates a tiny URL for file sharing
func ShareFile(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	if userID == nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	fileID := c.Param("id")
	var file File
	if err := DB.Where("unique_id = ? AND user_id = ?", fileID, userID).First(&file).Error; err != nil {
		c.HTML(http.StatusNotFound, "files.html", gin.H{"error": "File not found"})
		return
	}

	// Generate a tiny URL (for simplicity, using unique ID)
	tinyURL := fmt.Sprintf("http://localhost:8080/download/%s", file.UniqueID)

	c.HTML(http.StatusOK, "share.html", gin.H{"tinyURL": tinyURL})
}

// Logout user
func LogoutUser(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()

	c.Redirect(http.StatusSeeOther, "/login")
}
