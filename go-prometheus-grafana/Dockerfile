FROM golang:alpine@sha256:4db4aac30880b978cae5445dd4a706215249ad4f43d28bd7cdf7906e9be8dd6b as build
RUN mkdir /app
COPY . /app
WORKDIR /app
RUN go build -o main .

EXPOSE 2112
EXPOSE 8080

CMD ["/app/main"]