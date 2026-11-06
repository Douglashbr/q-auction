import { NestFactory } from '@nestjs/core';
import { ConfigService } from '@nestjs/config';
import { NextFunction, Request, Response } from 'express';
import { hostname } from 'node:os';
import { AppModule } from './app.module';
import { registerProxyRoutes } from './proxy/proxy';

async function bootstrap() {
  // Sem body parser: o gateway repassa o corpo das requisições intacto para os serviços.
  const app = await NestFactory.create(AppModule, { bodyParser: false });
  app.enableShutdownHooks();

  const instance = hostname();
  app.use((_req: Request, res: Response, next: NextFunction) => {
    res.setHeader('X-Instance', instance);
    next();
  });

  const config = app.get(ConfigService);
  registerProxyRoutes(app, config.getOrThrow<Record<string, string>>('routes'));

  await app.listen(config.getOrThrow<number>('port'));
}
bootstrap();
