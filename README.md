# Go Work Horse

Engine assíncrono de jobs em Go, com foco em performance, escalabilidade e confiabilidade.

## Visão Arquitetural

```
Producer → Queue → Worker Pool → Processor
```

- **Producer**: Envia jobs para a fila.
- **Queue**: Armazena jobs de forma plugável (Redis ou PostgreSQL).
- **Worker Pool**: Conjunto de workers que consomem jobs da fila.
- **Processor**: Executa a lógica de processamento do job.

Veja o diagrama detalhado em `docs/architecture.drawio`.

## Estrutura do Projeto

- `cmd/worker`: Entrypoint do worker.
- `pkg/jobqueue`: Lógica da fila e estrutura do job.
- `internal/repository`: Implementações de persistência.
- `configs/`: Configurações do projeto.
- `docs/`: Documentação e diagramas.

## Plugabilidade da Fila

A fila pode ser implementada usando Redis ou PostgreSQL. A escolha é feita via configuração.

## Estrutura de um Job

- `id`: Identificador único
- `payload`: Dados do job
- `status`: Status atual (pending, running, success, failed)
- `retry_count`: Número de tentativas
- `executed_at`: Data/hora da execução
- `created_at`: Data/hora de criação
- `updated_at`: Data/hora de atualização

## Configuração por ambiente

Crie um arquivo `.env` baseado em `.env.example` para configurar:

- `REDIS_ADDR`: endereço do Redis
- `WORKER_COUNT`: número de workers concorrentes
- `JOB_MAX_RETRIES`: tentativas máximas por job
- `JOB_RETRY_DELAY`: delay base para retries (segundos)
- `SIMULATE_FAIL`: simula falha para testar retries (1 = sim)

## Observabilidade

O sistema expõe métricas Prometheus em `/metrics` (porta 2112 por padrão).

Exemplo de scrape_config no Prometheus:

```yaml
scrape_configs:
  - job_name: "go_work_horse"
    static_configs:
      - targets: ["localhost:2112"]
```

Métricas expostas:

- `jobs_enqueued`: jobs atualmente na fila
- `jobs_processed_total`: jobs processados com sucesso
- `jobs_failed_total`: jobs com erro
- `jobs_retried_total`: jobs reprocessados

### Tracing

O sistema está instrumentado com OpenTelemetry (pontos principais: dequeue, processamento, retry). Configure um collector OTLP ou Jaeger para visualizar traces.

### Dashboard Grafana

Você pode importar um dashboard Prometheus/Grafana para visualizar as métricas. Exemplo de painel:

```json
{
  "title": "Go Work Horse Jobs",
  "panels": [
    {
      "type": "stat",
      "title": "Jobs Enqueued",
      "targets": [{ "expr": "jobs_enqueued" }]
    },
    {
      "type": "stat",
      "title": "Jobs Processed",
      "targets": [{ "expr": "jobs_processed_total" }]
    },
    {
      "type": "stat",
      "title": "Jobs Failed",
      "targets": [{ "expr": "jobs_failed_total" }]
    },
    {
      "type": "stat",
      "title": "Jobs Retried",
      "targets": [{ "expr": "jobs_retried_total" }]
    }
  ]
}
```

## Como rodar

Em breve.
