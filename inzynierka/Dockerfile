FROM golang:alpine

WORKDIR /app

# Instalacja Chromium, curl i czcionek dla renderowania wykresów PDF (chromedp)
RUN apk add --no-cache chromium font-noto ttf-freefont curl

# Instalacja Air oraz Delve (Debugger)
RUN go install github.com/air-verse/air@latest
RUN go install github.com/go-delve/delve/cmd/dlv@latest

COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Usuwamy stary CMD ["air"], bo przeniesiemy to do docker-compose
