# Build stage
FROM golang:1.19.3 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /app/config-syncer

# Final stage
FROM gcr.io/distroless/base
WORKDIR /app
COPY --from=build /app/config-syncer /app/config-syncer
ENTRYPOINT ["/app/config-syncer"]