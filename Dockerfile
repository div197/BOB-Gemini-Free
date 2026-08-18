FROM golang:alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

ARG VERSION=v0.1.0
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X main.Version=${VERSION}" -o /bob-gemini-free .

FROM alpine:3.21
RUN apk --no-cache add ca-certificates tzdata \
    && addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=builder /bob-gemini-free /app/bob-gemini-free

USER appuser

EXPOSE 8081

ENV BOB_GEMINI_FREE_HOST=0.0.0.0 \
    BOB_GEMINI_FREE_PORT=8081

ENTRYPOINT ["/app/bob-gemini-free"]
