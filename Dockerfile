FROM golang:1.20-alpine AS builder

WORKDIR /usr/local/src/tsrq

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/tsrq cmd/server/main.go

FROM alpine AS runner

COPY --from=builder /usr/local/bin/tsrq /usr/local/bin/tsrq

WORKDIR /usr/local/bin
CMD [ "tsrq" ]