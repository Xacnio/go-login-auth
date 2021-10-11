package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go-login-auth/database"
	"go-login-auth/routes"
	"os"
)

func init() {
	/* Load env variables from file */
	err := godotenv.Load(".env")
	if err != nil {
		panic(".env file couldn't be loaded: " + err.Error())
	}
	if os.Getenv("SC_KEY") == "" {
		panic("Define SC_KEY environment variable")
	}
}

func main() {
	/* Redis Connection */
	database.ConnectRedis()

	/* Database Connection */
	err := database.ConnectPSQL()
	if err != nil {
		panic(err)
	}
	defer database.DB.Close()

	/* Create API - HTTP Server */
	app := fiber.New()
	routes.Setup(app)
	host, port := os.Getenv("API_HOST"), os.Getenv("API_PORT")
	errHost := app.Listen(host + ":" + port)
	if errHost != nil {
		panic(errHost)
	}
}
