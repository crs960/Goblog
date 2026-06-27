package routes

import (
	"errors"
	"log"

	"goblog/internal/config"
	"goblog/internal/controllers"
	"goblog/internal/httpx"
	"goblog/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Setup(app *fiber.App, db *pgxpool.Pool, cfg config.Config) {
	authController := controllers.AuthController{DB: db, JWTSecret: cfg.JWTSecret}
	usersController := controllers.UsersController{DB: db}
	groupsController := controllers.GroupsController{DB: db}
	postsController := controllers.PostsController{DB: db}

	api := app.Group("/api")
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	authRoutes := api.Group("/auth")
	authRoutes.Post("/register", authController.Register)
	authRoutes.Post("/login", authController.Login)

	private := api.Use(middleware.JWT(cfg.JWTSecret))

	private.Get("/users/me", usersController.Me)
	private.Put("/users/me", usersController.UpdateMe)

	private.Post("/groups", groupsController.Create)
	private.Get("/groups", groupsController.List)
	private.Get("/groups/:id", groupsController.Get)
	private.Post("/groups/:id/join", groupsController.Join)
	private.Delete("/groups/:id/leave", groupsController.Leave)
	private.Get("/groups/:id/posts", postsController.ListByGroup)

	private.Post("/posts", postsController.Create)
	private.Get("/posts/:id", postsController.Get)
	private.Put("/posts/:id", postsController.Update)
	private.Delete("/posts/:id", postsController.Delete)
	private.Post("/posts/:id/like", postsController.Like)
	private.Delete("/posts/:id/like", postsController.Unlike)
	private.Post("/posts/:id/comments", postsController.CreateComment)
	private.Get("/posts/:id/comments", postsController.ListComments)
}

func ErrorHandler(c *fiber.Ctx, err error) error {
	var appErr httpx.AppError
	if errors.As(err, &appErr) {
		return c.Status(appErr.Status).JSON(fiber.Map{"erro": appErr.Message})
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return c.Status(fiberErr.Code).JSON(fiber.Map{"erro": fiberErr.Message})
	}

	log.Printf("erro interno: %v", err)
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"erro": "erro interno"})
}
