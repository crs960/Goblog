package main

import (
	"log"

	"goblog/internal/config"
	"goblog/internal/database"
	"goblog/internal/routes"

	"github.com/gofiber/fiber/v2"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("erro ao conectar no banco: %v", err)
	}
	defer db.Close()

	app := fiber.New(fiber.Config{
		AppName:      "goBlog API",
		ErrorHandler: routes.ErrorHandler,
	})

	routes.Setup(app, db, cfg)

	log.Printf("servidor rodando na porta %s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}
