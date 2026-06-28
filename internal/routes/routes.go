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
	usuariosController := controllers.UsuariosController{DB: db}
	gruposController := controllers.GruposController{DB: db}
	postsController := controllers.PostsController{DB: db}

	api := app.Group("/api")
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	authRoutes := api.Group("/auth")
	authRoutes.Post("/cadastro", authController.Cadastro)
	authRoutes.Post("/login", authController.Login)

	private := api.Use(middleware.JWT(cfg.JWTSecret))

	private.Get("/usuarios/eu", usuariosController.Eu)
	private.Put("/usuarios/eu", usuariosController.UpdateMe)

	private.Post("/grupo", gruposController.Create)
	private.Get("/grupo", gruposController.List)
	private.Get("/grupo/:id", gruposController.Get)
	private.Post("/grupo/:id/entrar", gruposController.Entrar)
	private.Delete("/grupo/:id/sair", gruposController.Sair)
	private.Get("/grupo/:id/post", postsController.ListByGrupo)

	private.Post("/post", postsController.Create)
	private.Get("/post/:id", postsController.Get)
	private.Put("/post/:id", postsController.Update)
	private.Delete("/post/:id", postsController.Delete)
	private.Post("/post/:id/like", postsController.Like)
	private.Delete("/post/:id/like", postsController.Unlike)
	private.Post("/post/:id/comentarios", postsController.CreateComentario)
	private.Get("/post/:id/comentarios", postsController.ListComentarios)
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