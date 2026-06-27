package controllers

import (
	"context"
	"errors"
	"strings"
	"time"

	"goblog/internal/auth"
	"goblog/internal/httpx"
	"goblog/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthController struct {
	DB        *pgxpool.Pool
	JWTSecret string
}

type registerRequest struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Senha string `json:"senha"`
}

type loginRequest struct {
	Email string `json:"email"`
	Senha string `json:"senha"`
}

func (h AuthController) Register(c *fiber.Ctx) error {
	var body registerRequest
	if err := c.BodyParser(&body); err != nil {
		return httpx.BadRequest("json invalido")
	}

	body.Nome = strings.TrimSpace(body.Nome)
	body.Email = strings.TrimSpace(strings.ToLower(body.Email))
	if body.Nome == "" || body.Email == "" || body.Senha == "" {
		return httpx.BadRequest("nome, email e senha sao obrigatorios")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	var usuario models.Usuario
	err := h.DB.QueryRow(ctx, `
		INSERT INTO usuarios (nome, email, senha_hash)
		VALUES ($1, $2, $3)
		RETURNING id, nome, email, senha_hash, criado_em, atualizado_em
	`, body.Nome, body.Email, body.Senha).Scan(
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

	token, err := auth.GenerateToken(usuario.ID, h.JWTSecret)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(models.AuthResponse{
		Token:   token,
		Usuario: usuario,
	})
}

func (h AuthController) Login(c *fiber.Ctx) error {
	var body loginRequest
	if err := c.BodyParser(&body); err != nil {
		return httpx.BadRequest("json invalido")
	}

	body.Email = strings.TrimSpace(strings.ToLower(body.Email))
	if body.Email == "" || body.Senha == "" {
		return httpx.BadRequest("email e senha sao obrigatorios")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	var usuario models.Usuario
	err := h.DB.QueryRow(ctx, `
		SELECT id, nome, email, senha_hash, criado_em, atualizado_em
		FROM usuarios
		WHERE email = $1
	`, body.Email).Scan(
		&usuario.ID,
		&usuario.Nome,
		&usuario.Email,
		&usuario.SenhaHash,
		&usuario.CriadoEm,
		&usuario.AtualizadoEm,
	)
	if errors.Is(err, pgx.ErrNoRows) || usuario.SenhaHash != body.Senha {
		return httpx.Unauthorized("email ou senha invalidos")
	}
	if err != nil {
		return err
	}

	token, err := auth.GenerateToken(usuario.ID, h.JWTSecret)
	if err != nil {
		return err
	}

	return c.JSON(models.AuthResponse{
		Token:   token,
		Usuario: usuario,
	})
}
