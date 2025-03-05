FROM golang:1.23 AS builder
WORKDIR /src

COPY go.* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/compass ./...

FROM alpine:latest
COPY --from=builder /bin/compass /bin/compass
CMD ["/bin/compass"]