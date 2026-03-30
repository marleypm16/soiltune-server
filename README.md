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
- `DBINFLUX`: URL do InfluxDB (ex: http://localhost:8086)
- `DBINFLUXTOKEN`: Token de autenticação do InfluxDB
- `DBINFLUXORG`: Organização do InfluxDB
- `DBINFLUXBUCKET`: Bucket do InfluxDB

## Como Executar

1. Configure as variáveis de ambiente necessárias.
2. Compile o projeto:
	 ```sh
	 go build -o soiltune-server
	 ```
3. Execute o binário:
	 ```sh
	 ./soiltune-server
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


