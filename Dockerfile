FROM golang:1.24 AS builder
WORKDIR /build

# cache deps
COPY go.mod go.sum ./
RUN go mod download

# build binary
COPY cmd cmd
COPY config config
COPY handler handler
COPY plugin plugin
RUN CGO_ENABLED=0 go build -o anygate ./cmd/anygate/main.go

FROM scratch

USER 1000
COPY --chown=1000:1000 --from=builder /build/anygate /anygate
EXPOSE 80
CMD ["/anygate"]
