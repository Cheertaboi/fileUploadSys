package main

import (
	"crypto/rand"
	"fmt"
	"net/http"
	
	"net/smtp"
	"github.com/gin-gonic/gin"
)

var otpStore = make(map[string]string)

func generateRandomOTP(length int) string {
	charset := "0123456789"
	b := make([]byte, length)
	rand.Read(b)
	otp := make([]byte, length)
	for i := range otp {
		otp[i] = charset[int(b[i])%len(charset)]
	}
	return string(otp)
}

func sendOTPEmail(toEmail, otp string) error {
	smtpHost := "email-smtp.eu-north-1.amazonaws.com"
	smtpPort := "587"
	smtpUser := "AKIA4IM3HJAE2AUBRTNP"
	smtpPass := "BD/rGW78wx65VJzuDp7aoQ/kEVYKTNDDuO7l8WG00Gjd"
	from := "harsh133hv@gmail.com"
	subject := "Your OTP for Password Reset"
	body := fmt.Sprintf("Your OTP is: %s", otp)

	msg := []byte("To: " + toEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" + body + "\r\n")
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	return smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{toEmail}, msg)
}

func main() {
	r := gin.Default()

	r.POST("/send-otp", func(c *gin.Context) {
		email := c.PostForm("email")
		if email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
			return
		}
		otp := generateRandomOTP(6)
		otpStore[email] = otp
		if err := sendOTPEmail(email, otp); err != nil {
			fmt.Println("Error sending email:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "OTP sent successfully"})
	})

	r.POST("/reset-password", func(c *gin.Context) {
		email := c.PostForm("email")
		token := c.PostForm("token")
		newPassword := c.PostForm("new_password")

		if storedToken, ok := otpStore[email]; !ok || token != storedToken {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OTP"})
			return
		}

		// Here, you would reset the password in the database
		fmt.Printf("Password for %s has been reset to: %s\n", email, newPassword)
		delete(otpStore, email)
		c.JSON(http.StatusOK, gin.H{"message": "Password reset successful"})
	})

	r.Run(":9090")
}
