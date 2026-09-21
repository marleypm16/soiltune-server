FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY api ./api
COPY consumer ./consumer
COPY internal ./internal

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/soiltune-api ./api && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/soiltune-consumer ./consumer

FROM alpine:3.22 AS runtime

RUN apk add --no-cache ca-certificates && \
    addgroup -S soiltune && \
    adduser -S -G soiltune soiltune

USER soiltune

FROM runtime AS api
COPY --from=build --chown=soiltune:soiltune /out/soiltune-api /usr/local/bin/soiltune-api
EXPOSE 8000
ENTRYPOINT ["/usr/local/bin/soiltune-api"]

FROM runtime AS consumer
COPY --from=build --chown=soiltune:soiltune /out/soiltune-consumer /usr/local/bin/soiltune-consumer
ENTRYPOINT ["/usr/local/bin/soiltune-consumer"]
