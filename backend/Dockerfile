FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY .env .env

RUN go build -o main .

EXPOSE ${PORT}

CMD ["./main"]