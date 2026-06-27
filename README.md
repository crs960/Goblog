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

- `POST /api/auth/register`
- `POST /api/auth/login`
- `GET /api/users/me`
- `PUT /api/users/me`
- `POST /api/groups`
- `GET /api/groups`
- `GET /api/groups/:id`
- `POST /api/groups/:id/join`
- `DELETE /api/groups/:id/leave`
- `GET /api/groups/:id/posts`
- `POST /api/posts`
- `GET /api/posts/:id`
- `PUT /api/posts/:id`
- `DELETE /api/posts/:id`
- `POST /api/posts/:id/like`
- `DELETE /api/posts/:id/like`
- `POST /api/posts/:id/comments`
- `GET /api/posts/:id/comments`