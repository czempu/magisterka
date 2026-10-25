FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod ./
COPY main.go ./
COPY index.html ./

RUN CGO_ENABLED=0 GOOS=linux go build -o rcp_client .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/rcp_client .

CMD ["./rcp_client"]
