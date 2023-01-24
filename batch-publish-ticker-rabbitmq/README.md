# Purpose
The idea of this repo is to experiment on how to
- batch publish data to RabbitMQ
- consume data from RabbitMQ until certain number of data fulfilled or after certain period of time

# Why
Sometimes you don't want to process data one by one. 

So initially you will store data into the DB after you consume it and create another cron to process the data.

This time I tried to do the process without storing it into the DB.

