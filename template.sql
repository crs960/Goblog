CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE usuarios (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    nome VARCHAR(100) NOT NULL,
    email VARCHAR(150) UNIQUE NOT NULL,
    senha_hash TEXT NOT NULL,

    criado_em TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    atualizado_em TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE grupos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    nome VARCHAR(150) NOT NULL,
    descricao TEXT,

    dono_id UUID NOT NULL,

    criado_em TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    atualizado_em TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_grupo_dono
        FOREIGN KEY (dono_id)
        REFERENCES usuarios(id)
        ON DELETE CASCADE
);

CREATE TABLE usuarios_grupos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    usuario_id UUID NOT NULL,
    grupo_id UUID NOT NULL,

    entrou_em TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_usuario_grupo_usuario
        FOREIGN KEY (usuario_id)
        REFERENCES usuarios(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_usuario_grupo_grupo
        FOREIGN KEY (grupo_id)
        REFERENCES grupos(id)
        ON DELETE CASCADE,

    CONSTRAINT unique_usuario_grupo
        UNIQUE (usuario_id, grupo_id)
);

CREATE TABLE postagens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    usuario_id UUID NOT NULL,

    grupo_id UUID NOT NULL,

    titulo VARCHAR(255) NOT NULL,
    conteudo TEXT,

    criado_em TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    atualizado_em TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_postagem_usuario
        FOREIGN KEY (usuario_id)
        REFERENCES usuarios(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_postagem_grupo
        FOREIGN KEY (grupo_id)
        REFERENCES grupos(id)
        ON DELETE CASCADE
);

CREATE TABLE curtidas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    usuario_id UUID NOT NULL,
    postagem_id UUID NOT NULL,

    curtido_em TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_curtida_usuario
        FOREIGN KEY (usuario_id)
        REFERENCES usuarios(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_curtida_postagem
        FOREIGN KEY (postagem_id)
        REFERENCES postagens(id)
        ON DELETE CASCADE,

    CONSTRAINT unique_curtida
        UNIQUE (usuario_id, postagem_id)
);

CREATE TABLE comentarios (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    usuario_id UUID NOT NULL,
    postagem_id UUID NOT NULL,

    comentario TEXT NOT NULL,

    criado_em TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    atualizado_em TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_comentario_usuario
        FOREIGN KEY (usuario_id)
        REFERENCES usuarios(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_comentario_postagem
        FOREIGN KEY (postagem_id)
        REFERENCES postagens(id)
        ON DELETE CASCADE
);


CREATE INDEX idx_grupos_dono
ON grupos(dono_id);

CREATE INDEX idx_usuarios_grupos_usuario
ON usuarios_grupos(usuario_id);

CREATE INDEX idx_usuarios_grupos_grupo
ON usuarios_grupos(grupo_id);

CREATE INDEX idx_postagens_usuario
ON postagens(usuario_id);

CREATE INDEX idx_postagens_grupo
ON postagens(grupo_id);

CREATE INDEX idx_curtidas_usuario
ON curtidas(usuario_id);

CREATE INDEX idx_curtidas_postagem
ON curtidas(postagem_id);

CREATE INDEX idx_comentarios_usuario
ON comentarios(usuario_id);

CREATE INDEX idx_comentarios_postagem
ON comentarios(postagem_id);

-- Trigger

CREATE OR REPLACE FUNCTION atualizar_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.atualizado_em = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_usuarios_updated
BEFORE UPDATE ON usuarios
FOR EACH ROW
EXECUTE FUNCTION atualizar_updated_at();

CREATE TRIGGER trg_grupos_updated
BEFORE UPDATE ON grupos
FOR EACH ROW
EXECUTE FUNCTION atualizar_updated_at();

CREATE TRIGGER trg_postagens_updated
BEFORE UPDATE ON postagens
FOR EACH ROW
EXECUTE FUNCTION atualizar_updated_at();

CREATE TRIGGER trg_comentarios_updated
BEFORE UPDATE ON comentarios
FOR EACH ROW
EXECUTE FUNCTION atualizar_updated_at();