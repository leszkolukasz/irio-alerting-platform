# API

This is the API microservice that serves publicly available endpoints.

## Running

API requires a running database, Redis and Pub/Sub instance. One can start them locally using Docker Compose:

```bash
docker compose up -d db redis
```

See Logger for Pub/Sub setup instructions.

To run the API service locally, use the following command:

```bash
go run .
```

## gRPC testing

To test the gRPC endpoints, you can use [Evans](https://github.com/ktr0731/evans).