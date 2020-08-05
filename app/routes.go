package app

import (
	v1controllers "authserver/controllers/v1"
	"authserver/middleware/auth"

	"github.com/gin-gonic/gin"
)

// SetupRoutes is the versioning entry point for routing
func SetupRoutes() {
	// Api Version 1
	v1 := api.Group("/v1")
	groupV1Routes(v1)
}

// groupV1Routes contains all of the grouped V1 routes
func groupV1Routes(v1 *gin.RouterGroup) {

	// status
	v1.GET("/heartbeat", v1controllers.Heartbeat)

	// groups
	admin := v1.Group("/admin")

	admin.POST("/login", v1controllers.AdminLogin)
	admin.POST("/refresh", v1controllers.AdminRefreshToken)
	admin.POST("/logout", auth.Admin(), v1controllers.AdminLogout)

	admin.GET("/super", auth.SuperAdmin(), v1controllers.Heartbeat)
	admin.GET("/admin", auth.Admin(), v1controllers.Heartbeat)
}
