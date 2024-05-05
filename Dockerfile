FROM golang:1.22.1

WORKDIR /app

COPY go.mod go.sum main.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /doers

EXPOSE 8080

CMD [ "/doers" ]