# AGENTS.md

Guia para **qualquer agente de IA** (Claude Code, Cursor, Codex, Copilot, Gemini...) e para humanos que contribuem com o q-auction. Siga estas regras em toda alteração.

Cada serviço tem também o próprio `AGENTS.md` (ex.: `services/user/AGENTS.md`), com regras da linguagem e comandos. **Leia o `AGENTS.md` do serviço antes de editá-lo.** Em caso de conflito, o arquivo mais próximo do código vale para aquele código, mas nunca pode contrariar as **regras inegociáveis** abaixo.

---

## 1. Visão geral

Auction house em microserviços, projeto de estudo. Cada serviço pode ser escrito numa linguagem diferente, e todos seguem **arquitetura hexagonal**.

```
host:80 → api-gateway (1x) ──/api/<svc>/*──→ <svc>-lb (nginx, round robin) → <svc>-service ×N (N ≥ 2)
               │                                                                  │
               └──────────── redis                       postgres · redis · kurrentdb
```

| Componente     | Onde                         | Porta no host | Papel                                               |
|----------------|------------------------------|---------------|-----------------------------------------------------|
| `api-gateway`  | `gateway/` (NestJS)          | 80            | Entrada única; roteia `/api/<svc>` para `<svc>-lb`  |
| `<svc>-lb`     | `infra/lb/` (nginx template) | —             | Load balancer do serviço (round robin)              |
| `<svc>-service`| `services/<svc>/`            | —             | Microserviço de domínio                             |
| `postgres`     | `postgres:17-alpine`         | 5433          | Um database + usuário por serviço                   |
| `redis`        | `redis:7.4-alpine`           | 6380          | Cache                                               |
| `kurrentdb`    | KurrentDB 26.0 LTS           | 2113          | Event store (UI em http://localhost:2113)           |

Serviços existentes:

| Serviço | Linguagem      | Rota no gateway | Doc                      |
|---------|----------------|-----------------|--------------------------|
| user    | Go 1.26 + Fiber v3 | `/api/user` | `services/user/AGENTS.md` |

## 2. Mapa do repositório

```
docker-compose.yml               orquestração de tudo
.env.example                     variáveis (copiar para .env; .env nunca é commitado)
gateway/                         API Gateway NestJS (fora da hexagonal, ver gateway/AGENTS.md)
services/<svc>/                  um diretório por microserviço, autocontido (código, Dockerfile, AGENTS.md)
infra/lb/default.conf.template   template nginx usado por TODOS os <svc>-lb
infra/postgres/init/             script que cria <svc>_db e <svc>_user
```

## 3. Regras inegociáveis

1. **Todo microserviço segue a arquitetura hexagonal** descrita na seção 4.
2. **Cada serviço tem seu próprio load balancer** (`<svc>-lb`) e **no mínimo 2 réplicas**. Instâncias de serviço **nunca publicam porta no host**.
3. **Cada serviço é dono dos seus dados.** Nunca leia nem escreva no database, nas chaves Redis ou nos streams de outro serviço.
4. **O gateway não tem regra de negócio.** Ele só roteia (e, no futuro, cuida de autenticação, rate limit e cache de borda).
5. **Versões fixadas.** Imagens Docker e dependências com versão exata ou travada em lockfile. Nunca use `latest`.
6. **Configuração só por variável de ambiente.** Nada de segredo no código. Novas variáveis entram no `.env.example`.
7. **Escopo.** Altere só o que a tarefa pede. Não mexa em outro serviço, na infra compartilhada ou no gateway sem que a tarefa exija; se for preciso, explique o porquê.

## 4. Arquitetura hexagonal

### Regra de dependência

As dependências apontam **sempre para dentro**:

```
adapters (in/out)  ──→  application (casos de uso + ports)  ──→  domain
        ↑
composition root (main/bootstrap) monta tudo
```

- `domain` não depende de nada (só da biblioteca padrão da linguagem).
- `application` depende só de `domain`.
- `adapters` dependem de `application` e de `domain`, nunca o contrário.
- O **composition root** (`main`) é o único lugar que instancia adapters e os injeta nos casos de uso.

### Estrutura canônica

Os nomes das camadas são os mesmos em todas as linguagens. Cada `AGENTS.md` de serviço mostra como isso fica na linguagem dele.

```
<src>/                     (internal/ em Go, src/ em TS/Java/Python, etc.)
  domain/                  entidades, value objects, eventos de domínio, erros de domínio
  application/             casos de uso (orquestram o domínio)
    ports/                 interfaces
      in                   o que o mundo externo pode pedir ao serviço (casos de uso)
      out                  o que o caso de uso precisa do mundo externo (repositório, event store, cache, publisher)
  adapters/
    in/                    http, consumidores de eventos, CLI — chamam os ports de entrada
    out/                   postgres, redis, kurrentdb, clientes HTTP de outros serviços — implementam os ports de saída
  config/                  leitura de variáveis de ambiente
main / bootstrap           composition root
```

### O que vai em cada camada

| Camada        | Pode conter                                                                 | Não pode conter                                        |
|---------------|------------------------------------------------------------------------------|--------------------------------------------------------|
| `domain`      | Regras de negócio, invariantes, entidades, value objects, eventos, erros    | Framework web, ORM, driver, JSON tags de API, logger   |
| `application` | Casos de uso, transações lógicas, interfaces (ports)                         | HTTP, SQL, detalhes de serialização, tipos de framework|
| `adapters/in` | Handlers, DTOs de request/response, validação de formato, mapeamento de erro | Regra de negócio                                       |
| `adapters/out`| Repositórios, clientes, modelos de persistência, mapeamento domínio↔tecnologia | Regra de negócio, chamadas a outros adapters         |

### Regras

- **Mapeamento na borda:** DTOs e modelos de banco nunca entram no domínio. O adapter converte para e de tipos do domínio.
- **Um adapter não chama outro adapter.** A comunicação passa sempre por um caso de uso.
- **Erros de domínio são tipados** (ex.: `ErrUserNotFound`). O adapter de entrada é o único que os traduz para status HTTP.
- **Nomenclatura dos ports:**
  - saída: `<Entidade>Repository`, `<Algo>Publisher`, `<Algo>Cache`, `<Servico>Client`;
  - entrada: `<Ação><Entidade>` / `<Ação>UseCase` (ex.: `RegisterUser`, `PlaceBid`).
- **Testabilidade:** domínio e casos de uso são testados **sem infraestrutura**, com fakes dos ports de saída.
- Não crie pacotes genéricos como `utils`, `helpers` ou `common`. Coloque o código na camada a que ele pertence.

## 5. Contrato de todo serviço

Todo serviço, em qualquer linguagem, deve:

- **Escutar na porta `PORT`** (padrão `3000`).
- **Expor `GET /health`** respondendo `200` com:
  ```json
  { "status": "ok", "service": "<svc>", "instance": "<hostname>" }
  ```
  `instance` é o hostname do container e serve para ver o round robin do LB.
- **Definir rotas sem o prefixo do gateway.** O gateway remove `/api/<svc>`, então `GET /api/user/health` chega ao serviço como `GET /health`.
- **Responder erros em JSON** neste formato (com `code` em snake_case):
  ```json
  { "error": { "code": "not_found", "message": "user not found" } }
  ```
- **Logar em stdout/stderr**, nunca em arquivo.
- **Encerrar de forma limpa no `SIGTERM`**, terminando as requisições em andamento.
- **Dockerfile** multi-stage, rodando com **usuário não-root** e com imagem base de versão fixa. A imagem final precisa ter `wget` (ou equivalente) para o healthcheck.
- **Healthcheck no compose** usando `127.0.0.1`, e não `localhost` (ver seção 10).

## 6. Dados e comunicação

| Recurso      | Convenção                                                                                       |
|--------------|-------------------------------------------------------------------------------------------------|
| Postgres     | Database `<svc>_db`, usuário `<svc>_user`, senha `<SVC>_DB_PASSWORD` no `.env`. Criados em `infra/postgres/init/01-create-databases.sh`. |
| Migrations   | Ficam dentro do serviço (`services/<svc>/migrations/`). O serviço aplica no startup, ou o `AGENTS.md` dele documenta o comando. |
| Redis        | Toda chave com prefixo `<svc>:` (ex.: `user:session:123`). Sempre com TTL quando for cache.  |
| KurrentDB    | Stream `<agregado>-<id>` (ex.: `auction-42`). Tipo de evento em PascalCase no passado (`AuctionCreated`, `BidPlaced`). Payload JSON em camelCase. Um serviço só escreve nos streams dos próprios agregados. |
| Entre serviços | **Prefira eventos** (KurrentDB). Se for preciso uma chamada síncrona, use `http://<outro>-lb` (o serviço entra na rede `services`). Nunca chame uma instância direto. |

Redes do compose:

- `services`: gateway ↔ LBs;
- `<svc>`: `<svc>-lb` ↔ instâncias do serviço;
- `data`: quem usa banco ou cache ↔ postgres/redis/kurrentdb.

Coloque o serviço em `data` somente quando ele passar a usar algum recurso de dados.

## 7. Adicionando um novo serviço

Exemplo com o serviço `bidding`:

1. Crie `services/bidding/` com a estrutura da seção 4, `GET /health`, `Dockerfile` e `AGENTS.md` (copie o formato de `services/user/AGENTS.md`).
2. No `docker-compose.yml`, adicione um bloco `# ---------- bidding ----------`:
   ```yaml
   bidding-lb:
     image: nginx:1.27-alpine
     environment:
       UPSTREAM: bidding-service:3000
     volumes:
       - ./infra/lb:/etc/nginx/templates:ro
     networks: [services, bidding]
     depends_on:
       bidding-service:
         condition: service_healthy
     healthcheck:
       test: ["CMD", "wget", "-qO-", "http://127.0.0.1/lb-health"]
       interval: 10s
       timeout: 3s
       retries: 5
     restart: unless-stopped

   bidding-service:
     build: ./services/bidding
     deploy:
       replicas: ${BIDDING_REPLICAS:-2}
     environment:
       PORT: 3000
     networks: [bidding]          # + data, se usar banco/cache
     healthcheck:
       test: ["CMD", "wget", "-qO-", "http://127.0.0.1:3000/health"]
       interval: 10s
       timeout: 3s
       retries: 5
       start_period: 5s
     restart: unless-stopped
   ```
   Depois, declare a rede `bidding:` em `networks`.
3. No `api-gateway`, adicione `ROUTES_BIDDING: http://bidding-lb`. Não é preciso mexer no código do gateway.
4. Se o serviço usa Postgres, adicione `create_db bidding_db bidding_user "$BIDDING_DB_PASSWORD"` no init script, a senha no `.env.example` e no `environment` do `postgres`. O init só roda com o volume vazio (ver seção 10).
5. Adicione as variáveis novas (`BIDDING_REPLICAS`...) no `.env.example`.
6. Atualize a tabela de serviços da seção 1 deste arquivo e o `README.md`.

## 8. Testes

- **Domínio e casos de uso:** testes unitários, sem rede, banco nem containers, usando fakes dos ports de saída.
- **Adapters de entrada:** testes do handler HTTP, com o caso de uso fake.
- **Adapters de saída:** testes de integração contra a infra real (Postgres, Redis, KurrentDB do compose).
- Todo serviço tem **um comando único** que roda os testes unitários, documentado no seu `AGENTS.md`.
- Toda regra de negócio nova precisa de teste no domínio ou no caso de uso.

## 9. Fluxo de trabalho para agentes

Antes de editar:

1. Leia este arquivo e o `AGENTS.md` do serviço ou componente que vai alterar.
2. Siga os padrões do código vizinho (o `services/user` é a referência de estrutura).

Antes de concluir:

1. Rode build, lint e testes do serviço (comandos no `AGENTS.md` dele).
2. Suba o serviço e valide pelo gateway:
   ```sh
   docker compose up -d --build <svc>-service
   docker compose ps                                   # tudo healthy
   for i in $(seq 1 4); do curl -s localhost/api/<svc>/health; echo; done   # instances alternando
   ```
3. Se mudou um contrato (rotas, eventos, variáveis, portas), atualize o `README.md` e os `AGENTS.md` afetados.

Git:

- Use **Conventional Commits** com o serviço como escopo: `feat(user): cadastra usuário`, `fix(gateway): ...`, `chore(infra): ...`, `docs: ...`.
- Faça commits pequenos e focados, com um serviço por commit sempre que possível.
- Nunca commite `.env`, `node_modules/`, `dist/` nem binários.

Nunca faça sem pedir:

- `docker compose down -v` ou `docker volume rm` (apagam todos os dados);
- mudar portas publicadas no host, versões de imagens de infra ou o template do LB;
- adicionar uma dependência nova sem dizer por quê.

## 10. Pegadinhas conhecidas

- **`localhost` em healthcheck:** em imagens alpine, `localhost` resolve primeiro para `::1` (IPv6), mas Fiber e nginx escutam só em IPv4. Use `127.0.0.1`.
- **Init do Postgres:** `infra/postgres/init/*.sh` só roda com o volume `pg-data` vazio. Para um database novo num ambiente que já existe, crie manualmente (`docker compose exec postgres psql -U admin -d postgres`) ou, com autorização, recrie o volume. O script precisa de permissão de execução (`chmod +x`).
- **Template do LB:** `infra/lb/default.conf.template` passa pelo `envsubst`. Só variáveis definidas no `environment` do container são substituídas; as do nginx (`$host`, `$remote_addr`) ficam intactas. Não defina env vars com nomes de variáveis do nginx.
- **Escala:** o LB reconsulta o DNS a cada 10s (`resolve`), então `docker compose up -d --scale <svc>-service=3` funciona sem reiniciar o LB.
- **Gateway:** o lockfile foi gerado com npm 11, então a imagem precisa ser `node:24`.
- **Portas no host:** Postgres em **5433** e Redis em **6380**, para não conflitar com instâncias locais.
