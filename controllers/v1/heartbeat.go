package v1controllers

import (
	"authserver/database"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Status struct {
	Database string `json:"database"`
	Message  string `json:"message"`
	Success  bool   `json:"success"`
}

// Heartbeat godoc
// @Summary ping example
// @Description do ping
// @Tags API Health
// @Accept json
// @Produce json
// @Success 200 {object} Status{database=string,message=string,success=bool}
// @Failure 500 {object} Status
// @Router /heartbeat [get]
func Heartbeat(c *gin.Context) {
	var databaseStatus string
	if err := database.Client.Ping(); err != nil {
		response := &Status{
			Success:  false,
			Message:  "Database inactive",
			Database: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := &Status{
		Success:  true,
		Message:  "All services active",
		Database: databaseStatus,
	}
	c.JSON(http.StatusOK, response)
}
