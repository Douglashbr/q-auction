# api-gateway — AGENTS.md

Porta de entrada única do sistema. Complementa o `AGENTS.md` da raiz, que deve ser lido primeiro.

- **Stack:** NestJS 11 + TypeScript, Node 24 (a imagem precisa ser `node:24-alpine` porque o lockfile foi gerado com npm 11).
- **Instâncias:** 1, publicada no host na porta 80 (`GATEWAY_PORT`).
- **O gateway é um componente de borda, não um microserviço de domínio.** Por isso fica **fora da arquitetura hexagonal** e usa a organização padrão do Nest (um módulo por responsabilidade).

## Regras

- **Sem regra de negócio.** Se a lógica depende do domínio de leilão, usuário ou lance, ela pertence a um serviço.
- **Não acessa Postgres nem KurrentDB.** Só usa Redis, para preocupações de borda (cache, rate limit, sessão).
- **Rotas para serviços vêm de configuração, não de código.** `ROUTES_<SVC>=http://<svc>-lb` no `environment` do compose gera `/api/<svc>/*`, com o prefixo removido antes de encaminhar (`src/proxy/proxy.ts`).
- **Sempre aponte para o LB do serviço** (`<svc>-lb`), nunca para `<svc>-service`. O gateway nem está na rede das instâncias.
- O body parser fica desligado (`bodyParser: false` em `main.ts`) para o corpo chegar intacto aos serviços. Endpoints próprios do gateway que precisem ler o body devem habilitar o parser só para eles.
- Toda resposta leva o header `X-Instance`.

## Estrutura

```
src/
  main.ts                     bootstrap, header X-Instance, registro do proxy
  app.module.ts
  config/configuration.ts     env → config (PORT, REDIS_URL, ROUTES_*)
  health/                     GET /health (terminus + indicador Redis)
  redis/redis.module.ts       provider ioredis global (token REDIS)
  proxy/proxy.ts              /api/<svc> → ROUTES_<SVC>
```

## Comandos

Execute dentro de `gateway/`:

```sh
npm ci
npm run build
REDIS_URL=redis://localhost:6380 ROUTES_USER=http://localhost:3001 npm run start:dev
```

No Docker, a partir da raiz: `docker compose up -d --build api-gateway`, depois `curl -s localhost/health`.
