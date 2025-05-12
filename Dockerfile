FROM golang:1.24 AS builder
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o app ./cmd/app/main.go

FROM scratch

USER 1000

COPY --chown=1000:1000 --from=builder /build/app /app

EXPOSE 8000
CMD ["/app"]