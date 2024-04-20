# Go RPC RabbitMQ

This is a demonstration where an API publish to rabbitmq, consumed by another consumer which will publish a callback to the API who wait to consume the data.

The API has
- endpoint /rpc
- publisher : publish to rpc_queue
- consumer 

The API will
- consume : from rpc_queue