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

type PostsController struct {
	DB *pgxpool.Pool
}

type postRequest struct {
	GrupoID  string  `json:"grupo_id"`
	Titulo   string  `json:"titulo"`
	Conteudo *string `json:"conteudo"`
}

type updatePostRequest struct {
	Titulo   string  `json:"titulo"`
	Conteudo *string `json:"conteudo"`
}

type commentRequest struct {
	Comentario string `json:"comentario"`
}

func (h PostsController) Create(c *fiber.Ctx) error {
	var body postRequest
	if err := c.BodyParser(&body); err != nil {
		return httpx.BadRequest("json invalido")
	}

	body.GrupoID = strings.TrimSpace(body.GrupoID)
	body.Titulo = strings.TrimSpace(body.Titulo)
	if body.GrupoID == "" || body.Titulo == "" {
		return httpx.BadRequest("grupo_id e titulo sao obrigatorios")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	ok, err := repository.UsuarioNoGrupo(ctx, h.DB, middleware.UserID(c), body.GrupoID)
	if err != nil {
		return err
	}
	if !ok {
		return httpx.Forbidden("voce precisa entrar no grupo antes de postar")
	}

	var postagem models.Postagem
	err = h.DB.QueryRow(ctx, `
		INSERT INTO postagens (usuario_id, grupo_id, titulo, conteudo)
		VALUES ($1, $2, $3, $4)
		RETURNING id, usuario_id, grupo_id, titulo, conteudo, 0, 0, false, criado_em, atualizado_em
	`, middleware.UserID(c), body.GrupoID, body.Titulo, body.Conteudo).Scan(postScanArgs(&postagem)...)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(postagem)
}

func (h PostsController) ListByGroup(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	ok, err := repository.UsuarioNoGrupo(ctx, h.DB, middleware.UserID(c), c.Params("id"))
	if err != nil {
		return err
	}
	if !ok {
		return httpx.Forbidden("voce nao participa deste grupo")
	}

	rows, err := h.DB.Query(ctx, postSelectQuery(`
		WHERE p.grupo_id = $1
		ORDER BY p.criado_em DESC
	`), c.Params("id"), middleware.UserID(c))
	if err != nil {
		return err
	}
	defer rows.Close()

	postagens := []models.Postagem{}
	for rows.Next() {
		var postagem models.Postagem
		if err := rows.Scan(postScanArgs(&postagem)...); err != nil {
			return err
		}
		postagens = append(postagens, postagem)
	}

	return c.JSON(postagens)
}

func (h PostsController) Get(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	ok, err := repository.UsuarioPodeVerPostagem(ctx, h.DB, middleware.UserID(c), c.Params("id"))
	if err != nil {
		return err
	}
	if !ok {
		return httpx.Forbidden("voce nao tem acesso a esta postagem")
	}

	var postagem models.Postagem
	err = h.DB.QueryRow(ctx, postSelectQuery(`
		WHERE p.id = $1
	`), c.Params("id"), middleware.UserID(c)).Scan(postScanArgs(&postagem)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return httpx.NotFound("postagem nao encontrada")
	}
	if err != nil {
		return err
	}

	return c.JSON(postagem)
}

func (h PostsController) Update(c *fiber.Ctx) error {
	var body updatePostRequest
	if err := c.BodyParser(&body); err != nil {
		return httpx.BadRequest("json invalido")
	}

	body.Titulo = strings.TrimSpace(body.Titulo)
	if body.Titulo == "" {
		return httpx.BadRequest("titulo e obrigatorio")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	var postagem models.Postagem
	err := h.DB.QueryRow(ctx, postSelectQuery(`
		WHERE p.id = $1 AND p.usuario_id = $3
	`), c.Params("id"), middleware.UserID(c), middleware.UserID(c)).Scan(postScanArgs(&postagem)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return httpx.NotFound("postagem nao encontrada para este usuario")
	}
	if err != nil {
		return err
	}

	err = h.DB.QueryRow(ctx, `
		WITH updated AS (
			UPDATE postagens
			SET titulo = $3, conteudo = $4
			WHERE id = $1 AND usuario_id = $2
			RETURNING id
		)
		SELECT
			p.id,
			p.usuario_id,
			p.grupo_id,
			p.titulo,
			p.conteudo,
			COUNT(DISTINCT l.id)::int AS total_curtidas,
			COUNT(DISTINCT c.id)::int AS total_comentarios,
			COALESCE(BOOL_OR(l.usuario_id = $2), false) AS curtida_pelo_usuario,
			p.criado_em,
			p.atualizado_em
		FROM postagens p
		INNER JOIN updated u ON u.id = p.id
		LEFT JOIN curtidas l ON l.postagem_id = p.id
		LEFT JOIN comentarios c ON c.postagem_id = p.id
		GROUP BY p.id
	`, c.Params("id"), middleware.UserID(c), body.Titulo, body.Conteudo).Scan(postScanArgs(&postagem)...)
	if err != nil {
		return err
	}

	return c.JSON(postagem)
}

func (h PostsController) Delete(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	tag, err := h.DB.Exec(ctx, `
		DELETE FROM postagens
		WHERE id = $1 AND usuario_id = $2
	`, c.Params("id"), middleware.UserID(c))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httpx.NotFound("postagem nao encontrada para este usuario")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h PostsController) Like(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	ok, err := repository.UsuarioPodeVerPostagem(ctx, h.DB, middleware.UserID(c), c.Params("id"))
	if err != nil {
		return err
	}
	if !ok {
		return httpx.Forbidden("voce nao tem acesso a esta postagem")
	}

	_, err = h.DB.Exec(ctx, `
		INSERT INTO curtidas (usuario_id, postagem_id)
		VALUES ($1, $2)
	`, middleware.UserID(c), c.Params("id"))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return httpx.Conflict("postagem ja curtida")
		}

		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h PostsController) Unlike(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	tag, err := h.DB.Exec(ctx, `
		DELETE FROM curtidas
		WHERE usuario_id = $1 AND postagem_id = $2
	`, middleware.UserID(c), c.Params("id"))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httpx.NotFound("curtida nao encontrada")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h PostsController) CreateComment(c *fiber.Ctx) error {
	var body commentRequest
	if err := c.BodyParser(&body); err != nil {
		return httpx.BadRequest("json invalido")
	}

	body.Comentario = strings.TrimSpace(body.Comentario)
	if body.Comentario == "" {
		return httpx.BadRequest("comentario e obrigatorio")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	ok, err := repository.UsuarioPodeVerPostagem(ctx, h.DB, middleware.UserID(c), c.Params("id"))
	if err != nil {
		return err
	}
	if !ok {
		return httpx.Forbidden("voce nao tem acesso a esta postagem")
	}

	var comentario models.Comentario
	err = h.DB.QueryRow(ctx, `
		INSERT INTO comentarios (usuario_id, postagem_id, comentario)
		VALUES ($1, $2, $3)
		RETURNING id, usuario_id, postagem_id, comentario, criado_em, atualizado_em
	`, middleware.UserID(c), c.Params("id"), body.Comentario).Scan(
		&comentario.ID,
		&comentario.UsuarioID,
		&comentario.PostagemID,
		&comentario.Comentario,
		&comentario.CriadoEm,
		&comentario.AtualizadoEm,
	)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(comentario)
}

func (h PostsController) ListComments(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	ok, err := repository.UsuarioPodeVerPostagem(ctx, h.DB, middleware.UserID(c), c.Params("id"))
	if err != nil {
		return err
	}
	if !ok {
		return httpx.Forbidden("voce nao tem acesso a esta postagem")
	}

	rows, err := h.DB.Query(ctx, `
		SELECT id, usuario_id, postagem_id, comentario, criado_em, atualizado_em
		FROM comentarios
		WHERE postagem_id = $1
		ORDER BY criado_em ASC
	`, c.Params("id"))
	if err != nil {
		return err
	}
	defer rows.Close()

	comentarios := []models.Comentario{}
	for rows.Next() {
		var comentario models.Comentario
		if err := rows.Scan(
			&comentario.ID,
			&comentario.UsuarioID,
			&comentario.PostagemID,
			&comentario.Comentario,
			&comentario.CriadoEm,
			&comentario.AtualizadoEm,
		); err != nil {
			return err
		}
		comentarios = append(comentarios, comentario)
	}

	return c.JSON(comentarios)
}

func postSelectQuery(where string) string {
	return `
		SELECT
			p.id,
			p.usuario_id,
			p.grupo_id,
			p.titulo,
			p.conteudo,
			COUNT(DISTINCT l.id)::int AS total_curtidas,
			COUNT(DISTINCT c.id)::int AS total_comentarios,
			COALESCE(BOOL_OR(l.usuario_id = $2), false) AS curtida_pelo_usuario,
			p.criado_em,
			p.atualizado_em
		FROM postagens p
		LEFT JOIN curtidas l ON l.postagem_id = p.id
		LEFT JOIN comentarios c ON c.postagem_id = p.id
		` + where + `
		GROUP BY p.id
	`
}

func postScanArgs(postagem *models.Postagem) []any {
	return []any{
		&postagem.ID,
		&postagem.UsuarioID,
		&postagem.GrupoID,
		&postagem.Titulo,
		&postagem.Conteudo,
		&postagem.TotalCurtidas,
		&postagem.TotalComentarios,
		&postagem.CurtidaPeloUsuario,
		&postagem.CriadoEm,
		&postagem.AtualizadoEm,
	}
}
