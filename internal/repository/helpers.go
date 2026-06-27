package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func UsuarioNoGrupo(ctx context.Context, db *pgxpool.Pool, usuarioID, grupoID string) (bool, error) {
	var exists bool
	err := db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM usuarios_grupos
			WHERE usuario_id = $1 AND grupo_id = $2
		)
	`, usuarioID, grupoID).Scan(&exists)
	return exists, err
}

func UsuarioPodeVerPostagem(ctx context.Context, db *pgxpool.Pool, usuarioID, postagemID string) (bool, error) {
	var exists bool
	err := db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM postagens p
			INNER JOIN usuarios_grupos ug ON ug.grupo_id = p.grupo_id
			WHERE p.id = $1 AND ug.usuario_id = $2
		)
	`, postagemID, usuarioID).Scan(&exists)
	return exists, err
}
