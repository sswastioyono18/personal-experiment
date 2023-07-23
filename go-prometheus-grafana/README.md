# Purpose

I want to learn how to create docker app with grafana, prometheus with golang
The objective is to pull metric from golang application and display it in grafana


# How to run (Docker)
Just run `docker-compose up -d` 
- Hit application in `localhost:2112/metrics` to see custom metrics `http_response_status`
- Search prometheus metric from grafana in `localhost:3000` with status code `200` or `400`'
