FROM golang:tip-alpine3.22 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Build the Go app
RUN CGO_ENABLED=0 go build -o /2do ./cmd/2do

FROM scratch
WORKDIR /app
COPY migrations/ ./migrations/
COPY --from=builder /2do ./2do
EXPOSE 8080
CMD ["/app/2do"]
