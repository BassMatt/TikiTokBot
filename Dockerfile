FROM golang:1.16-alpine


WORKDIR /app

COPY . .

RUN go build -o /tikitokbot

CMD ["/tikitokbot"]