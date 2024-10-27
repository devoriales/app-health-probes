FROM golang:1.21.6-alpine as builder

WORKDIR /app

COPY . .

RUN sed -i 's/Critical application/Critical app running in container/g' main.go


RUN CGO_ENABLED=0 go build -o main .


FROM alpine:latest  
WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
