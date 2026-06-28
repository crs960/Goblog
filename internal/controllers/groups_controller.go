package controllers

import (
	"context"
	"errors"
	"strings"
	"time"

	"goblog/internal/httpx"
	"goblog/internal/middleware"
	"goblog/internal/models"
	"goblog/internal/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GruposController struct {
	DB *pgxpool.Pool
}

type grupoRequest struct {
	Nome      string  `json:"nome"`
	Descricao *string `json:"descricao"`
}

func (h GruposController) Create(c *fiber.Ctx) error {
	var body grupoRequest
	if err := c.BodyParser(&body); err != nil {
		return httpx.BadRequest("json invalido")
	}

	body.Nome = strings.TrimSpace(body.Nome)
	if body.Nome == "" {
		return httpx.BadRequest("nome e obrigatorio")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	tx, err := h.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var grupo models.Grupo
	err = tx.QueryRow(ctx, `
		INSERT INTO grupos (nome, descricao, dono_id)
		VALUES ($1, $2, $3)
		RETURNING id, nome, descricao, dono_id, criado_em, atualizado_em
	`, body.Nome, body.Descricao, middleware.UserID(c)).Scan(
		&grupo.ID,
		&grupo.Nome,
		&grupo.Descricao,
		&grupo.DonoID,
		&grupo.CriadoEm,
		&grupo.AtualizadoEm,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO usuarios_grupos (usuario_id, grupo_id)
		VALUES ($1, $2)
	`, middleware.UserID(c), grupo.ID)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(grupo)
}

func (h GruposController) List(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	rows, err := h.DB.Query(ctx, `
		SELECT g.id, g.nome, g.descricao, g.dono_id, g.criado_em, g.atualizado_em
		FROM grupos g
		INNER JOIN usuarios_grupos ug ON ug.grupo_id = g.id
		WHERE ug.usuario_id = $1
		ORDER BY g.criado_em DESC
	`, middleware.UserID(c))
	if err != nil {
		return err
	}
	defer rows.Close()

	grupos := []models.Grupo{}
	for rows.Next() {
		var grupo models.Grupo
		if err := rows.Scan(&grupo.ID, &grupo.Nome, &grupo.Descricao, &grupo.DonoID, &grupo.CriadoEm, &grupo.AtualizadoEm); err != nil {
			return err
		}
		grupos = append(grupos, grupo)
	}

	return c.JSON(grupos)
}

func (h GruposController) Get(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	ok, err := repository.UsuarioNoGrupo(ctx, h.DB, middleware.UserID(c), c.Params("id"))
	if err != nil {
		return err
	}
	if !ok {
		return httpx.Forbidden("voce nao participa deste grupo")
	}

	var grupo models.Grupo
	err = h.DB.QueryRow(ctx, `
		SELECT id, nome, descricao, dono_id, criado_em, atualizado_em
		FROM grupos
		WHERE id = $1
	`, c.Params("id")).Scan(&grupo.ID, &grupo.Nome, &grupo.Descricao, &grupo.DonoID, &grupo.CriadoEm, &grupo.AtualizadoEm)
	if errors.Is(err, pgx.ErrNoRows) {
		return httpx.NotFound("grupo nao encontrado")
	}
	if err != nil {
		return err
	}

	return c.JSON(grupo)
}

func (h GruposController) Entrar(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	_, err := h.DB.Exec(ctx, `
		INSERT INTO usuarios_grupos (usuario_id, grupo_id)
		VALUES ($1, $2)
	`, middleware.UserID(c), c.Params("id"))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return httpx.Conflict("usuario ja esta no grupo")
		}
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return httpx.NotFound("grupo nao encontrado")
		}

		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h GruposController) Sair(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	tag, err := h.DB.Exec(ctx, `
		DELETE FROM usuarios_grupos
		WHERE usuario_id = $1 AND grupo_id = $2
	`, middleware.UserID(c), c.Params("id"))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httpx.NotFound("participacao no grupo nao encontrada")
	}

	return c.SendStatus(fiber.StatusNoContent)
}
