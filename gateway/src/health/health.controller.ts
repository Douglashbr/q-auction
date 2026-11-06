import { Controller, Get } from '@nestjs/common';
import { HealthCheck, HealthCheckService } from '@nestjs/terminus';
import { hostname } from 'node:os';
import { RedisHealthIndicator } from './redis.health';

@Controller('health')
export class HealthController {
  constructor(
    private readonly health: HealthCheckService,
    private readonly redis: RedisHealthIndicator,
  ) {}

  @Get()
  @HealthCheck()
  async check() {
    const result = await this.health.check([() => this.redis.isHealthy('redis')]);
    // instance identifica a réplica que respondeu (útil para ver o round robin).
    return { ...result, instance: hostname() };
  }
}
