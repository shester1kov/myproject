FROM golang:1.23.1-alpine

WORKDIR /project

COPY go.mod go.sum ./

RUN go mod tidy

COPY . .

RUN go build -o myproject ./cmd

EXPOSE 8080

CMD ["./myproject"]
