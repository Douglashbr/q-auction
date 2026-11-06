import { INestApplication, Logger } from '@nestjs/common';
import { createProxyMiddleware } from 'http-proxy-middleware';

// Monta /api/<serviço> → <url>. O Express remove o prefixo montado,
// então /api/auction/items chega ao serviço como /items.
export function registerProxyRoutes(app: INestApplication, routes: Record<string, string>) {
  const logger = new Logger('Proxy');
  for (const [name, target] of Object.entries(routes)) {
    app.use(`/api/${name}`, createProxyMiddleware({ target, changeOrigin: true }));
    logger.log(`/api/${name} -> ${target}`);
  }
}
