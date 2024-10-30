# HomeBunny, your favourite Smart Home Assistant

HomeBunny is a system utilizing [Event Driven Architecture](https://aws.amazon.com/event-driven-architecture/) to create a Smart Home Assistant capable of monitoring and controlling various IoT devices in a home environment.

The devices will collect data and trigger actions based on user-defined thresholds or events.

## Prerequisities

- Docker https://www.docker.com/
- PostgreSQL https://www.postgresql.org/

## Database

The database schema is contained in `/backend/database.sql`.
Please run those commands in psql in order to create the database, replacing `your_user` and `your_password` with your own setup details.

## RabbitMQ Setup

Once docker has been installed, run the following command in your console:

`docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3.11-management`

The command will start RabbitMQ in a docker container, called rabbitmq.

Next, please create the user, virtual host, and topic exchange with the commands below.

In an instance where { } appear, replace with your input; for example, `docker exec rabbitmq rabbitmqctl add_user kay-hk s3cr3t`

```docker exec rabbitmq rabbitmqctl add_user {username} {password}
docker exec rabbitmq rabbitmqctl set_user_tags {username} administrator
docker exec rabbitmq rabbitmqctl delete_user guest
docker exec rabbitmq rabbitmqctl add_vhost customers
docker exec rabbitmq rabbitmqctl set_permissions -p customers {username} ".*" ".*" ".*"
docker exec rabbitmq rabbitmqadmin declare exchange --vhost=customers name=device_events type=topic -u {username} -p {password} durable=true
docker exec rabbitmq rabbitmqctl set_topic_permissions -p customers {username} device_events ".*" ".*"
```

## YAML configuration

In file `backend/internal/rabbitmq.go`, change the line 27 to the correct path (`../smart-home-assistant/backend/config/config.yaml`) in order to load in the config.

In `backend/config/config.yaml` input the values from your RabbitMQ and PostgreSQL setup in order for the app to be able to use the values.

# Running the application

`go run cmd/server/main.go`

`go run cmd/consumer/main.go`

`go run cmd/producer/main.go`

Each console will print information as events are pulished and consumed.

## Testing

```
go build cmd/server/main.go
go build cmd/producer/main.go
go build cmd/consumer/main.go
```

To run the test inside each directory:

`go test -v`

### Future Improvements

- RabbitMQ init.sh file
- Create database programatically in go
- Executable
- Integrate Prometheus and Grafana
- Frontend
