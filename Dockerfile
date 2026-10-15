# ========== ЭТАП 1: сборка ==========
# Берём официальный образ Go. Он большой (~800 МБ), но нужен только для сборки.
FROM golang:1.22-alpine AS builder

# Рабочая директория внутри контейнера.
WORKDIR /app

# Копируем файлы зависимостей СНАЧАЛА — чтобы Docker кэшировал их отдельно.
# Если менять только код — зависимости пересобираться не будут. Быстрее.
COPY go.mod go.sum ./

# Скачиваем зависимости (gorilla/websocket, modernc.org/sqlite, bcrypt).
RUN go mod download

# Теперь копируем ВЕСЬ остальной код (main.go, handler.go, static/ и т.д.).
COPY . .

# Собираем бинарник.
# CGO_ENABLED=0 — отключаем C-зависимости (важно для SQLite).
# GOOS=linux — под Linux (а не Windows/Mac, где ты разрабатываешь).
# -ldflags="-s -w" — убираем отладочную информацию. Бинарник ~30% меньше.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o myapp .

# ========== ЭТАП 2: финальный образ ==========
# Alpine — маленький образ Linux (~5 МБ).
FROM alpine:latest

# CA-сертификаты нужны, если приложение ходит в HTTPS (на будущее).
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Копируем ТОЛЬКО бинарник из этапа сборки. Исходники и Go-компилятор не нужны.
COPY --from=builder /app/myapp .

# Копируем статику (index.html).
COPY --from=builder /app/static ./static

# Говорим Docker: "приложение слушает порт 8080".
EXPOSE 8080

# Что запускать при старте контейнера.
CMD ["./myapp"]