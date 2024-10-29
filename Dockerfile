FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod ./
COPY *.go ./

RUN go build -o /clima_cep

EXPOSE 8080

CMD [ "/clima_cep" ]
