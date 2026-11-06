const ROUTE_PREFIX = 'ROUTES_';

// Lê as rotas dos serviços a partir de variáveis ROUTES_<NOME>=<url>.
// Ex.: ROUTES_AUCTION=http://auction-service:3000 → /api/auction/* é encaminhado para lá.
function parseRoutes(env: NodeJS.ProcessEnv): Record<string, string> {
  const routes: Record<string, string> = {};
  for (const [key, value] of Object.entries(env)) {
    if (key.startsWith(ROUTE_PREFIX) && value) {
      routes[key.slice(ROUTE_PREFIX.length).toLowerCase()] = value;
    }
  }
  return routes;
}

export default () => ({
  port: parseInt(process.env.PORT ?? '3000', 10),
  redisUrl: process.env.REDIS_URL ?? 'redis://localhost:6379',
  routes: parseRoutes(process.env),
});
