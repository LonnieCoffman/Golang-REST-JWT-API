package main

import (
	"authserver/app"

	"github.com/joho/godotenv"
)

func init() {
	// load local environment variables if they exist
	godotenv.Load()
}

// @securityDefinitions.apikey bearerAuth
// @in header
// @name Authorization
func main() {

	app.StartApplication()
}
