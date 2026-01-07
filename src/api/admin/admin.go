package api

import (
	"mixproxy/src/proxy/config"

	"github.com/gofiber/fiber/v2/middleware/cors"
)

func HandleAdminAPI() {
	api := config.SERVERS["HTTPS"].Group("/api")
	api.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,OPTIONS,DELETE",
		AllowHeaders: "Content-Type",
	}))

	// Public routes
	SetupLogsPublic(api)

	// Middleware
	api.Use(adminApiMiddleware)

	// Protected routes
	SetupLogsProtected(api)
	SetupControlRoutes(api)
	SetupStatsRoutes(api)
	SetupConfigRoutes(api)
	SetupWhitelistRoutes(api)
	SetupBlacklistRoutes(api)
	SetupDockerRoutes(api)
}
