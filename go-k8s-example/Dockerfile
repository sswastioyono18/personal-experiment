FROM golang:alpine as build
RUN mkdir /app
COPY . /app
WORKDIR /app
RUN go build -o main .

CMD ["/app/main"]