# goBlog API

## Rodando

1. Crie um banco de dados postgres

2. Execute os comandos em template.sql para criar as tabelas

3. Edite os valores conforme seu PostgreSQL:
```env
    # DB postgres
    DB_USER=postgres
    DB_PASSWORD=1234
    DB_NAME=postgres
    DB_HOST=localhost
    DB_PORTS=5432:5432
    DB_PORT=5432
    DB_SSLMODE=disable

    JWT_SECRET=goblog-secret
    PORT=3000
```

4. Inicie a API:

```bash
go run .
```

## Rotas principais

- `POST /api/auth/cadastro`
- `POST /api/auth/login`
- `GET /api/usuarios/eu`
- `PUT /api/usuarios/eu`
- `POST /api/grupo`
- `GET /api/grupo`
- `GET /api/grupo/:id`
- `POST /api/grupo/:id/entrar`
- `DELETE /api/grupo/:id/sair`
- `GET /api/grupo/:id/post`
- `POST /api/post`
- `GET /api/post/:id`
- `PUT /api/post/:id`
- `DELETE /api/post/:id`
- `POST /api/post/:id/like`
- `DELETE /api/post/:id/like`
- `POST /api/post/:id/comentarios`
- `GET /api/post/:id/comentarios`
