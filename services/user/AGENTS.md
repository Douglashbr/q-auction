# user-service — AGENTS.md

Serviço de usuários. Complementa o `AGENTS.md` da raiz, que deve ser lido primeiro.

- **Stack:** Go 1.26, Fiber v3. Módulo `q-auction/user`.
- **Rota no gateway:** `/api/user/*` → `user-lb` → `user-service` (2 réplicas).
- **Dados:** ainda não usa banco. Quando usar: `user_db` / `user_user` (senha `USER_DB_PASSWORD`), e o serviço entra na rede `data`.

## Comandos

Execute dentro de `services/user/`:

```sh
gofmt -l .                 # deve sair vazio
go vet ./...
go test ./...              # testes unitários (comando único do serviço)
PORT=3001 go run ./cmd/user
```

Dentro do Docker, a partir da raiz:

```sh
docker compose up -d --build user-service
curl -s localhost/api/user/health
```

## Estrutura (hexagonal em Go)

```
cmd/user/main.go                  composition root: config → adapters → casos de uso → Listen + shutdown
internal/
  config/                         Load() lê variáveis de ambiente
  domain/                         entidades, value objects, erros (ErrXxx), eventos. Só stdlib.
  application/                    casos de uso (um arquivo por caso de uso)
    ports/                        interfaces de entrada e saída (ex.: UserRepository)
  adapters/
    in/httpapi/                   Fiber: server.go (app + error handler), um arquivo por recurso, *_test.go
    out/<tecnologia>/             ex.: out/postgres, out/redis, out/kurrentdb
```

- O pacote HTTP se chama `httpapi`, e não `http`, para não colidir com `net/http`.
- `internal/` garante que nenhum outro módulo importe o código do serviço.

## Convenções Go

- O caso de uso é uma struct que recebe os ports de saída no construtor (`NewRegisterUser(repo ports.UserRepository)`) e expõe um método `Execute(ctx context.Context, input) (output, error)`.
- `context.Context` é sempre o primeiro parâmetro de qualquer operação que faça I/O.
- Erros de domínio são variáveis sentinela em `domain` (`var ErrUserNotFound = errors.New("user not found")`), comparadas com `errors.Is`. O mapeamento erro → status HTTP fica **só** no `errorHandler` de `adapters/in/httpapi/server.go`.
- Structs de domínio **sem** tags `json`. DTOs com tags ficam em `adapters/in/httpapi`.
- Handlers registram as próprias rotas (`func (h *XHandler) Register(app *fiber.App)`), e o `main` só chama `Register`.
- Fiber v3: handlers recebem `fiber.Ctx` (interface, sem ponteiro).
- Não use variáveis globais nem `init()` para montar dependências. Tudo é montado no `main`.
- Dependência nova só com justificativa, e sempre `go mod tidy` depois.

## Testes

- Handlers: `app.Test(httptest.NewRequest(...))` com o caso de uso fake (ver `internal/adapters/in/httpapi/health_test.go`).
- Casos de uso: fakes em memória dos ports de saída, definidos no próprio `_test.go`.
- Adapters de saída: testes de integração contra o compose, marcados com build tag `integration` (`go test -tags integration ./...`).
