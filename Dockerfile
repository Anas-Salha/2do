FROM golang:tip-alpine3.22 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY src/ ./src/

# Build the Go app
RUN CGO_ENABLED=0 go build -o /2do ./src/cmd/server

FROM scratch
WORKDIR /app
COPY --from=builder /2do /app/2do
CMD ["/app/2do"]
