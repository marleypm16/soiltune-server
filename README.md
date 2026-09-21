# soiltune-server

Servidor dedicado do projeto Soiltune para armazenar dados de sensores enviados por dispositivos ESP via MQTT e persistir essas informações no InfluxDB.

## Visão Geral

O soiltune-server atua como uma ponte entre dispositivos IoT (como ESP32/ESP8266) e um banco de dados InfluxDB. Ele recebe dados de sensores via MQTT, processa e armazena essas informações de forma eficiente para posterior análise.

## Funcionalidades

- **Recepção de Dados via MQTT:** O servidor se conecta a um broker MQTT e escuta mensagens em um tópico configurável.
- **Persistência no InfluxDB:** Cada mensagem recebida é convertida e armazenada como um ponto no InfluxDB, facilitando consultas temporais e análises.
- **Configuração via Variáveis de Ambiente:** Todos os parâmetros sensíveis (URLs, tokens, tópicos) são configurados por variáveis de ambiente, facilitando o deploy em diferentes ambientes.

## Estrutura dos Dados

O payload MQTT deve ser um JSON com o seguinte formato:

```json
{
	"sensor_id": "string",
	"temperature": 0.0,
	"humidity": 0.0,
	"weight": 0.0
}
```

## Variáveis de Ambiente

- `MQTTBROKER`: URL do broker MQTT (ex: tcp://localhost:1883)
- `MQTTTOPIC`: Tópico MQTT para inscrição
- `MQTT_USERNAME`: Usuário do backend no broker MQTT
- `MQTT_PASSWORD`: Senha do backend no broker MQTT
- `MQTT_DEVICE_ID`: Identificador e usuário MQTT do dispositivo local de demonstração
- `MQTT_DEVICE_PASSWORD`: Senha do dispositivo local de demonstração
- `API_KEY`: Chave com pelo menos 32 caracteres usada para autorizar comandos
- `DBINFLUX`: URL do InfluxDB (ex: http://localhost:8086)
- `DOCKER_INFLUXDB_INIT_ADMIN_TOKEN`: Token de autenticação do InfluxDB
- `DOCKER_INFLUXDB_INIT_ORG`: Organização do InfluxDB
- `DOCKER_INFLUXDB_INIT_BUCKET`: Bucket do InfluxDB

## Envio de comandos

A API envia apenas os comandos `0` (desligar) e `1` (ligar). O identificador da rota deve corresponder ao usuário MQTT do dispositivo.

```sh
curl -X POST http://localhost:8080/command/sensor-01 \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":1}'
```

Uma resposta `202 Accepted` confirma que o comando foi publicado no broker; ela não confirma que o dispositivo alterou seu estado.

O backend pode ler `soiltune/telemetry/#` e escrever em `soiltune/commands/#`. Cada dispositivo autenticado pode escrever apenas em `soiltune/telemetry/{device_id}` e ler apenas `soiltune/commands/{device_id}`.

As senhas MQTT devem ter pelo menos 16 caracteres. No Compose, a API fica vinculada a `127.0.0.1`; para exposição externa, use um proxy reverso com TLS. O listener MQTT local exige autenticação e ACL, mas ainda usa TCP sem criptografia: antes de disponibilizá-lo fora de uma rede confiável, configure MQTT sobre TLS.

## Como Executar

1. Configure as variáveis de ambiente necessárias.
2. Compile os dois executáveis:
	 ```sh
	 go build -o soiltune-api ./api
	 go build -o soiltune-consumer ./consumer
	 ```
3. Execute os binários:
	 ```sh
	 ./soiltune-consumer
	 ./soiltune-api
	 ```

## Dependências

- Go 1.26+
- [github.com/eclipse/paho.mqtt.golang](https://github.com/eclipse/paho.mqtt.golang)
- [github.com/influxdata/influxdb-client-go/v2](https://github.com/influxdata/influxdb-client-go)

## Estrutura do Projeto

- `main.go`: Ponto de entrada da aplicação.
- `mqtt.go`: Lida com a conexão e assinatura MQTT.
- `influxdb.go`: Processa e armazena os dados no InfluxDB.
- `go.mod`: Gerenciamento de dependências.

## Exemplo de Uso

Dispositivo ESP publica no tópico MQTT configurado:

```json
{
	"sensor_id": "esp32-01",
	"temperature": 23.5,
	"humidity": 60.2,
	"weight": 150.0
}
```

O servidor irá registrar automaticamente esses dados no InfluxDB.


