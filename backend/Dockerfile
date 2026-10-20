# Один образ на два бинаря: APP=api (сервер) или APP=mocks (моки внешних сервисов).
# Ожидаемая структура: ./cmd/api/main.go и ./cmd/mocks/main.go.

FROM golang:1.25-alpine AS build
WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
ARG APP=api
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/${APP}

FROM alpine:3.22
RUN adduser -D -u 10001 app
COPY --from=build /out/app /app
USER app
ENTRYPOINT ["/app"]
