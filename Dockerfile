# Build stage
FROM golang:1.23-bookworm AS builder

WORKDIR /app

COPY ./ /app

RUN go mod download
RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -o /azureadoexporter /app/cmd/azureadoexporter

# Final stage
FROM gcr.io/distroless/static-debian12

WORKDIR /

COPY --from=builder /azureadoexporter /azureadoexporter

# Create a non-root user
USER nonroot:nonroot

EXPOSE 8080

CMD ["/azureadoexporter"]