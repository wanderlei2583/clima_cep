FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /go/bin/clima-cep

FROM alpine:3.18

RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/clima-cep /clima-cep

RUN adduser -D appuser
USER appuser

EXPOSE 8080
CMD [ "/clima-cep" ]
