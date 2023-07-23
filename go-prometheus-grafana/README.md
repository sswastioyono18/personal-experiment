# Purpose

I want to learn how to create docker app with grafana, promtail and loki.

The objective is to be able to browse log produced by go application from grafana


# How to run (Docker)
Just run `docker-compose up -d` 
- Hit application in `localhost:8080`
- Search log from grafana in `localhost:3000`

# How to run (Kubernetes)
Just run `kubectl apply -f deployment.yaml`

Then you need to port forward grafana in port 3000