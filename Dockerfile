FROM golang:1.26-alpine AS base

WORKDIR /app



# Baixa as dependências do Go
COPY go.mod go.sum ./
RUN go mod download

# Copia todo o código fonte
COPY . .


# --- Target da API ---
FROM base AS api
RUN go build -o /api-bin ./api
EXPOSE 8080
CMD ["/api-bin"]


# --- Target do Consumer ---
FROM base AS consumer
# Compila e roda
RUN go build -o /consumer-bin ./consumer
CMD ["/consumer-bin"]