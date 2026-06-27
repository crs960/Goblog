package models

import "time"

type Usuario struct {
	ID           string    `json:"id"`
	Nome         string    `json:"nome"`
	Email        string    `json:"email"`
	SenhaHash    string    `json:"-"`
	CriadoEm     time.Time `json:"criado_em"`
	AtualizadoEm time.Time `json:"atualizado_em"`
}

type Grupo struct {
	ID           string    `json:"id"`
	Nome         string    `json:"nome"`
	Descricao    *string   `json:"descricao"`
	DonoID       string    `json:"dono_id"`
	CriadoEm     time.Time `json:"criado_em"`
	AtualizadoEm time.Time `json:"atualizado_em"`
}

type Postagem struct {
	ID                 string    `json:"id"`
	UsuarioID          string    `json:"usuario_id"`
	GrupoID            string    `json:"grupo_id"`
	Titulo             string    `json:"titulo"`
	Conteudo           *string   `json:"conteudo"`
	TotalCurtidas      int       `json:"total_curtidas"`
	TotalComentarios   int       `json:"total_comentarios"`
	CurtidaPeloUsuario bool      `json:"curtida_pelo_usuario"`
	CriadoEm           time.Time `json:"criado_em"`
	AtualizadoEm       time.Time `json:"atualizado_em"`
}

type Comentario struct {
	ID           string    `json:"id"`
	UsuarioID    string    `json:"usuario_id"`
	PostagemID   string    `json:"postagem_id"`
	Comentario   string    `json:"comentario"`
	CriadoEm     time.Time `json:"criado_em"`
	AtualizadoEm time.Time `json:"atualizado_em"`
}

type AuthResponse struct {
	Token   string  `json:"token"`
	Usuario Usuario `json:"usuario"`
}
