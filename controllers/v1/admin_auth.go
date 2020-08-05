package v1controllers

import (
	"authserver/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminLogin(c *gin.Context) {
	admin := models.Admin{}

	fmt.Printf("%T\n", models.Admin{})

	if err := c.ShouldBindJSON(&admin); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"success": false, "message": "Invalid JSON provided"})
		return
	}

	if status, err := admin.IsAuthenticated(); err != nil {
		c.JSON(status, gin.H{"success": false, "message": err.Error()})
		return
	}

	if err := admin.GetAuthToken(); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"success": false, "message": err.Error()})
		return
	}

	if err := admin.GetRefreshToken(); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"success": false, "message": err.Error()})
		return
	}

	admin.RemoteAddress = c.ClientIP()

	if status, err := admin.StoreSession(); err != nil {
		c.JSON(status, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  admin.AccessToken,
		"message":       "Authenticated",
		"refresh_token": admin.RefreshToken,
		"success":       true,
	})
}

// AdminLogout ...
func AdminLogout(c *gin.Context) {
	admin := models.Admin{}

	adminID, ok := c.Get("AdminID")
	if ok {
		admin.ID = adminID.(uint64)
	}

	adminToken, ok := c.Get("AdminToken")
	if ok {
		admin.AccessToken = adminToken.(string)
	}

	admin.Logout()

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out",
		"success": true,
	})
}

// AdminRefreshToken ...
func AdminRefreshToken(c *gin.Context) {
	admin := models.Admin{}

	if err := c.ShouldBindJSON(&admin); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"success": false, "message": "Invalid JSON provided"})
		return
	}

	if status, err := admin.ValidateRefreshToken(); err != nil {
		c.JSON(status, gin.H{"success": false, "message": err.Error()})
		return
	}

	if err := admin.GetAuthToken(); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"success": false, "message": err.Error()})
		return
	}

	if err := admin.GetRefreshToken(); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"success": false, "message": err.Error()})
		return
	}

	admin.RemoteAddress = c.ClientIP()

	if status, err := admin.UpdateSession(); err != nil {
		c.JSON(status, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  admin.AccessToken,
		"message":       "Authenticated",
		"refresh_token": admin.RefreshToken,
		"success":       true,
	})
}

// AdminResetPassword ...
func AdminResetPassword(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"Controller": "admin reset password"})
}
