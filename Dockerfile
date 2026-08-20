FROM golang:alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

ARG VERSION=v0.1.7
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X main.Version=${VERSION}" -o /bob-gemini-free .

FROM alpine:3.21
RUN apk --no-cache add ca-certificates tzdata \
    && addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=builder /bob-gemini-free /app/bob-gemini-free

USER appuser

EXPOSE 9610

ENV BOB_GEMINI_FREE_HOST=0.0.0.0 \
    BOB_GEMINI_FREE_PORT=9610

HEALTHCHECK --interval=20s --timeout=3s --start-period=3s --retries=3 \
  CMD wget -qO- http://127.0.0.1:9610/ >/dev/null 2>&1 || exit 1

ENTRYPOINT ["/app/bob-gemini-free"]
