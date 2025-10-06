package main

import (
	"day/config"
	"day/database"
	"day/routes"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
)

func init() {
	config.LoadEnv()
	database.InitDatabase()
}

func main() {
	r := gin.Default()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Static("/uploads", "./uploads")
	
	routes.InitRoutes(r)

	r.Run(fmt.Sprintf(":%s", port))
}
