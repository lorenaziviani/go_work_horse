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

## Como rodar

Em breve.
