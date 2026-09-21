# soiltune-server

Backend IoT do Soiltune para receber telemetria de dispositivos via MQTT, persistir séries temporais no InfluxDB e enviar comandos de ligar/desligar aos dispositivos.

## Arquitetura

```text
Dispositivo ── MQTT/QoS 1 ──► Mosquitto ──► Consumer ──► InfluxDB ──► Grafana
     ▲                              ▲
     └──── comando MQTT ◄──── API HTTP autenticada
```

O repositório gera dois executáveis:

- `soiltune-consumer`: valida mensagens MQTT e grava os pontos no InfluxDB;
- `soiltune-api`: recebe comandos HTTP autenticados e os publica no MQTT.

## Contrato de telemetria v1

O dispositivo deve publicar em `soiltune/telemetry/{device_id}`. O `sensor_id` do JSON deve ser igual ao `device_id` do tópico.

```json
{
  "version": 1,
  "sensor_id": "sensor-01",
  "recorded_at": "2026-09-21T12:30:00Z",
  "temperature": 23.5,
  "humidity": 60.2,
  "weight": 150.0,
  "state": "on"
}
```

Regras:

- `version` deve ser `1`;
- `sensor_id` deve começar com letra ou número e pode conter letras, números, `_` e `-`, até 64 caracteres;
- `recorded_at` é obrigatório e deve estar em RFC3339;
- `temperature`, `humidity` e `weight` são obrigatórios, inclusive quando o valor é zero;
- umidade deve estar entre 0 e 100;
- peso não pode ser negativo;
- `state` é opcional e aceita entre 1 e 32 caracteres;
- campos desconhecidos e múltiplos objetos JSON são rejeitados.

No InfluxDB, o ponto usa a measurement `sensor_data`, as tags `sensor_id` e `schema_version`, e o timestamp informado em `recorded_at`.

### Migração do protótipo

Mensagens antigas sem `version` e `recorded_at` passam a ser rejeitadas. Os campos do InfluxDB também foram padronizados de `temperatura`, `umidade`, `peso` e `estado` para `temperature`, `humidity`, `weight` e `state`. Pontos antigos permanecem no banco, mas consultas e dashboards devem considerar os novos nomes.

## API de comandos

### `POST /command/:sensorId`

Publica `0` para desligar ou `1` para ligar em `soiltune/commands/{sensorId}`.

```sh
curl -X POST http://localhost:8080/command/sensor-01 \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":1}'
```

Respostas principais:

- `202 Accepted`: o broker confirmou a publicação;
- `400 Bad Request`: identificador, JSON ou comando inválido;
- `401 Unauthorized`: chave ausente ou inválida;
- `503 Service Unavailable`: MQTT indisponível.

`202` não confirma que o dispositivo alterou fisicamente seu estado. Confirma apenas o despacho para o broker.

### Saúde

- `GET /health/live`: processo HTTP ativo;
- `GET /health/ready`: API conectada ao MQTT.

## Execução com Docker Compose

Pré-requisitos:

- Docker com Docker Compose;
- portas 1883, 3000, 8080 e 8086 disponíveis.

1. Crie a configuração local:

   ```sh
   cp .env.example .env
   ```

   No PowerShell:

   ```powershell
   Copy-Item .env.example .env
   ```

2. Preencha todos os valores vazios de `.env`. Use senhas MQTT com pelo menos 16 caracteres e uma `API_KEY` aleatória com pelo menos 32 caracteres.

3. Suba a stack:

   ```sh
   docker compose up --build
   ```

4. Verifique a API:

   ```sh
   curl http://localhost:8080/health/ready
   ```

O Compose aguarda InfluxDB e Mosquitto ficarem saudáveis antes de iniciar os processos dependentes e reinicia serviços interrompidos.

## Execução local sem Docker

É necessário fornecer InfluxDB e Mosquitto externamente e ajustar `DBINFLUX` e `MQTTBROKER`.

```sh
go build -o soiltune-api ./api
go build -o soiltune-consumer ./consumer

./soiltune-consumer
./soiltune-api
```

## Configuração

| Variável | Obrigatória | Padrão | Finalidade |
| --- | --- | --- | --- |
| `DBINFLUX` | sim | — | URL do InfluxDB |
| `DOCKER_INFLUXDB_INIT_MODE` | no Compose | `setup` no exemplo | Inicialização do InfluxDB |
| `DOCKER_INFLUXDB_INIT_USERNAME` | no Compose | — | Usuário administrativo inicial |
| `DOCKER_INFLUXDB_INIT_PASSWORD` | no Compose | — | Senha administrativa inicial |
| `DOCKER_INFLUXDB_INIT_ORG` | sim | `soiltune` no exemplo | Organização do InfluxDB |
| `DOCKER_INFLUXDB_INIT_BUCKET` | sim | `sensor-data` no exemplo | Bucket de telemetria |
| `DOCKER_INFLUXDB_INIT_ADMIN_TOKEN` | sim | — | Token do InfluxDB |
| `INFLUX_WRITE_TIMEOUT` | não | `5s` | Timeout por tentativa de escrita |
| `INFLUX_WRITE_ATTEMPTS` | não | `3` | Tentativas de escrita, entre 1 e 10 |
| `MQTTBROKER` | sim | — | URL do broker |
| `MQTTTOPIC` | no consumer | — | Filtro de telemetria |
| `MQTT_USERNAME` | sim | — | Usuário do backend |
| `MQTT_PASSWORD` | sim | — | Senha do backend |
| `MQTT_QOS` | não | `1` | QoS de publicação e assinatura |
| `MQTT_DEVICE_ID` | no Compose | — | Usuário/ID do dispositivo local |
| `MQTT_DEVICE_PASSWORD` | no Compose | — | Senha do dispositivo local |
| `API_KEY` | na API | — | Bearer token com pelo menos 32 caracteres |
| `API_PORT` | não | `8000` | Porta interna da API |

Variáveis do processo têm precedência sobre um arquivo `.env` local.

## Segurança MQTT

O broker local:

- não permite acesso anônimo;
- fornece ao backend leitura de toda a telemetria e escrita de comandos;
- permite a cada dispositivo escrever apenas em `soiltune/telemetry/{seu_usuario}`;
- permite a cada dispositivo ler apenas `soiltune/commands/{seu_usuario}`;
- limita payloads MQTT a 64 KiB.

A API fica vinculada a `127.0.0.1:8080`. Para exposição externa, use um proxy reverso com TLS.

O listener MQTT do Compose usa autenticação e ACL, mas ainda opera sem criptografia. Não exponha a porta 1883 à internet. Para produção, configure MQTT sobre TLS na porta 8883 ou use um broker gerenciado.

## Garantias de entrega

MQTT usa QoS 1 por padrão, portanto broker e cliente podem entregar duplicatas. O processamento deve continuar idempotente do ponto de vista do produto.

Uma escrita no InfluxDB possui timeout e retry limitado. Depois de esgotadas as tentativas, a mensagem é registrada como rejeitada e não existe fila durável local. O projeto não promete entrega exatamente uma vez.

## Testes e qualidade

```sh
go test ./...
go vet ./...
gofmt -w ./api ./consumer ./internal
```

Os testes protegem:

- autenticação da API;
- validação e despacho de comandos;
- contrato e mapeamento da telemetria;
- correspondência entre tópico e `sensor_id`;
- retry de escrita;
- configuração e health/readiness.

Com a stack Docker em execução, os dois fluxos externos podem ser verificados com:

```sh
go test -tags=integration ./integration
```

Esse teste publica telemetria MQTT e confirma sua presença no InfluxDB, além de enviar um comando HTTP e confirmar seu recebimento por um cliente MQTT autenticado.

## Estrutura

```text
api/                    API HTTP de comandos
consumer/               ingestão MQTT e escrita no InfluxDB
internal/config/        configuração e validação de ambiente
internal/models/        contratos de telemetria e comandos
internal/mosquitto/     configuração do broker local
docker-compose.yaml     stack local
Dockerfile              imagens da API e do consumer
```

## Limitações conhecidas

- o broker local ainda não possui TLS;
- `202 Accepted` não representa confirmação física do dispositivo;
- depois que o retry do InfluxDB é esgotado, a leitura é perdida;
- o Compose provisiona uma conta de dispositivo para demonstração; ambientes com vários dispositivos devem gerenciar credenciais individualmente.
