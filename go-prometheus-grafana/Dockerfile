FROM golang:alpine as build
RUN mkdir /app
COPY . /app
WORKDIR /app
RUN go build -o main .

EXPOSE 2112
EXPOSE 8080

CMD ["/app/main"]