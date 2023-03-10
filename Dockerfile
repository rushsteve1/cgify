FROM golang:alpine AS builder

WORKDIR /go/delivery

COPY go.mod .
COPY main.go .

RUN go build

# --- #

FROM alpine:edge AS app

COPY --from=builder /go/delivery/cgify /bin
