package main

import (
	"crud-app/config"
	"crud-app/database"
	"crud-app/route"
	"log"
)

func main() {
	config.LoadEnv()    // dari evn.go
	config.InitLogger() // dari logger.go

	database.ConnectDB()
	defer database.DB.Close()

	app := config.NewApp()

	route.SetupRoutes(app, database.DB)

	port := config.GetEnv("APP_PORT", "3000")

	config.Logger.Println("🚀 Server running at http://localhost:" + port)

	log.Fatal(app.Listen(":" + port))
}
