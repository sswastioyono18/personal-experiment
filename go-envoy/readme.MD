Experimenting running envoy as API gateway
- run docker compose up -d
- running app in local which is in port 8080
- envoy to route from port 10000 to 8080
- request from browser with localhost:10000
- log incoming request with timestamp will be showed in golang app

- In localhost, this request will send twice and its normal due to behaviour of the browser.

If you use postman, it will only log 1 request. 