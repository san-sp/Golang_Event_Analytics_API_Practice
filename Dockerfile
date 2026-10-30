
# Build stage
FROM golang:1.27.1 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /pulsetrack-api ./cmd/api

# Runtime stage
FROM gcr.io/distroless/static-debian13:nonroot

WORKDIR /

COPY --from=builder /pulsetrack-api /pulsetrack-api

EXPOSE 8100

USER nonroot:nonroot

ENTRYPOINT ["/pulsetrack-api"]
