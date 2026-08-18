FROM golang:alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

ARG VERSION=dev
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=${VERSION}" -o /bob-gemini-free .

FROM scratch
WORKDIR /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bob-gemini-free /app/bob-gemini-free
EXPOSE 8081

ENTRYPOINT ["/app/bob-gemini-free"]
