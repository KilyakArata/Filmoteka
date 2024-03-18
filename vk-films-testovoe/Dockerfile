FROM golang:1.21-alpine AS builder

WORKDIR /usr/local/src

RUN apk --no-cache add bash git make gcc gettext musl-dev

COPY ["go.mod", "go.sum", "./"]
RUN go mod download

COPY filmoteka ./
RUN go build -o ./cmd/main/ cmd/main/main.go

FROM alpine as runner

COPY --from=builder /usr/local/src/cmd/main/ /

CMD ["/main"]
