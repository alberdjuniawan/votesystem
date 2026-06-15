FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o votesystem ./cmd/api

FROM alpine:latest

RUN apk --no-cache add tzdata ca-certificates

RUN addgroup -S apps && adduser -S apps -G apps

WORKDIR /app

COPY --from=builder /app/votesystem .

RUN chown apps:apps /app/votesystem

USER apps

EXPOSE 8080

CMD ["./votesystem"]