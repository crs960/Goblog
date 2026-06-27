package controllers

import (
	"context"
	"errors"
	"strings"
	"time"

	"goblog/internal/httpx"
	"goblog/internal/middleware"
	"goblog/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UsersController struct {
	DB *pgxpool.Pool
}

type updateUserRequest struct {
	Nome  *string `json:"nome"`
	Email *string `json:"email"`
	Senha *string `json:"senha"`
}

func (h UsersController) Me(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	var usuario models.Usuario
	err := h.DB.QueryRow(ctx, `
		SELECT id, nome, email, senha_hash, criado_em, atualizado_em
		FROM usuarios
		WHERE id = $1
	`, middleware.UserID(c)).Scan(
		&usuario.ID,
		&usuario.Nome,
		&usuario.Email,
		&usuario.SenhaHash,
		&usuario.CriadoEm,
		&usuario.AtualizadoEm,
	)
	if err != nil {
		return err
	}

	return c.JSON(usuario)
}

func (h UsersController) UpdateMe(c *fiber.Ctx) error {
	var body updateUserRequest
	if err := c.BodyParser(&body); err != nil {
		return httpx.BadRequest("json invalido")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	var usuario models.Usuario
	err := h.DB.QueryRow(ctx, `
		UPDATE usuarios
		SET
			nome = COALESCE(NULLIF($2, ''), nome),
			email = COALESCE(NULLIF($3, ''), email),
			senha_hash = COALESCE(NULLIF($4, ''), senha_hash)
		WHERE id = $1
		RETURNING id, nome, email, senha_hash, criado_em, atualizado_em
	`,
		middleware.UserID(c),
		normalizeOptional(body.Nome),
		normalizeEmail(body.Email),
		normalizeOptional(body.Senha),
	).Scan(
		&usuario.ID,
		&usuario.Nome,
		&usuario.Email,
		&usuario.SenhaHash,
		&usuario.CriadoEm,
		&usuario.AtualizadoEm,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return httpx.Conflict("email ja cadastrado")
		}

		return err
	}

	return c.JSON(usuario)
}

func normalizeOptional(value *string) string {
	if value == nil {
		return ""
	}

	return strings.TrimSpace(*value)
}

func normalizeEmail(value *string) string {
	if value == nil {
		return ""
	}

	return strings.TrimSpace(strings.ToLower(*value))
}
