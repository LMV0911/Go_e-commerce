FROM golang:1.24-alpine AS builder

WORKDIR /app
ENV GOTOOLCHAIN=auto

RUN apk add --no-cache git

COPY . .
RUN go mod tidy && go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/admin ./cmd/admin
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/client ./cmd/client
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/ssh ./cmd/ssh

FROM alpine:latest

WORKDIR /app
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/bin/ /app/bin/

EXPOSE 8080 2222

CMD ["/app/bin/server"]
