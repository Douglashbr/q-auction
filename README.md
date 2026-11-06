# q-auction

Auction house em microserviços (projeto de estudo).

```
host:80 → api-gateway (NestJS, 1x) ──/api/user/*──→ user-lb (nginx, round robin) → user-service ×2 (Go + Fiber)
               │
       postgres · redis · kurrentdb
```

Cada serviço tem o **próprio load balancer** e no mínimo 2 instâncias. O gateway só conhece o LB de cada serviço, nunca as instâncias.

| Serviço      | Imagem / build                                 | Host               | Papel                                      |
|--------------|------------------------------------------------|--------------------|--------------------------------------------|
| api-gateway  | `./gateway` (NestJS)                           | `localhost:80`     | Porta de entrada, roteia `/api/<serviço>`  |
| user-lb      | `nginx:1.27-alpine` + `infra/lb`               | —                  | LB do serviço de usuários (round robin)    |
| user-service | `./services/user` (Go + Fiber), 2 réplicas     | —                  | Serviço de usuários                        |
| postgres     | `postgres:17-alpine`                           | `localhost:5433`   | Um database + usuário por serviço          |
| redis        | `redis:7.4-alpine`                             | `localhost:6380`   | Cache (AOF, LRU com 256mb)                 |
| kurrentdb    | `docker.kurrent.io/kurrent-lts/kurrentdb:26.0` | `localhost:2113`   | Event store (UI em http://localhost:2113)  |

Postgres e Redis usam 5433/6380 no host para não conflitar com instâncias locais. Dá para mudar no `.env`.

Redes:
- `services`: gateway ↔ load balancers
- `user`: user-lb ↔ instâncias do user-service
- `data`: quem precisa de banco/cache ↔ postgres, redis, kurrentdb

## Contribuindo (humanos e agentes de IA)

As regras do projeto (arquitetura hexagonal, contrato dos serviços, convenções de dados, como adicionar um serviço, commits) estão no **[AGENTS.md](AGENTS.md)**. Cada serviço tem o seu `AGENTS.md` com regras da linguagem e comandos (ex.: [services/user/AGENTS.md](services/user/AGENTS.md)).

Cursor, Codex, Copilot, Gemini e outros leem o `AGENTS.md` automaticamente. O Claude Code lê o `CLAUDE.md`, que só importa o `AGENTS.md`.

## Subindo

```sh
cp .env.example .env
docker compose up -d --build
docker compose ps        # tudo "healthy"
```

## Adicionando um novo serviço

O LB é um template genérico (`infra/lb/default.conf.template`). A imagem do nginx substitui `${UPSTREAM}` na inicialização. Para um serviço `bidding`:

1. Crie `services/bidding` com `GET /health` e um Dockerfile.
2. No compose, adicione:
   - `bidding-service` com `deploy.replicas: 2` na rede `bidding` (e em `data`, se usar banco);
   - `bidding-lb` igual ao `user-lb`, com `UPSTREAM: bidding-service:3000` e redes `services` + `bidding`;
   - a rede `bidding`.
3. No `api-gateway`, adicione `ROUTES_BIDDING: http://bidding-lb`.

## Endpoints

- `GET /health`: health do gateway (inclui Redis).
- `GET /api/user/health`: health do user-service, passando por gateway → user-lb. Retorna `instance` (hostname da réplica que atendeu).

## Verificação

```sh
# Round robin: os hostnames das 2 réplicas se alternam
for i in $(seq 1 6); do curl -s localhost/api/user/health | jq -r .instance; done
docker compose logs user-lb | grep upstream=          # qual instância atendeu

# Failover: pare uma réplica e repita o loop, sem erros
docker stop q-auction-user-service-1

# Escala sem reiniciar o LB (ele reconsulta o DNS a cada 10s)
docker compose up -d --scale user-service=3 user-service

# Postgres
docker compose exec postgres psql -U admin -d postgres -c '\l'

# Redis
docker compose exec redis redis-cli ping

# KurrentDB: grava e lê um evento
curl -i -X POST localhost:2113/streams/auction-test \
  -H 'Content-Type: application/vnd.eventstore.events+json' \
  -d "[{\"eventId\":\"$(uuidgen | tr A-Z a-z)\",\"eventType\":\"AuctionCreated\",\"data\":{\"title\":\"teste\"}}]"
curl -s localhost:2113/streams/auction-test/0 -H 'Accept: application/json'
```

## Desenvolvimento local

```sh
cd gateway && npm install && REDIS_URL=redis://localhost:6380 ROUTES_USER=http://localhost:3001 npm run start:dev
cd services/user && PORT=3001 go run ./cmd/user
```

## Observações

- O script `infra/postgres/init/01-create-databases.sh` só roda com o volume vazio. Para recriar: `docker compose down -v` (apaga **todos** os dados).
- Os healthchecks usam `127.0.0.1` e não `localhost`: no alpine, `localhost` resolve primeiro para `::1`, mas o Fiber e o nginx escutam só em IPv4.
- KurrentDB roda em modo inseguro (sem TLS/autenticação), apenas para desenvolvimento.
