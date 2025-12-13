# API

This is the API microservice that serves publicly available endpoints.

## Running

API requires a running database and redis instance. One can start them locally using Docker Compose:

```bash
docker compose up -d db redis
```

To run the API service locally, use the following command:

```bash
go run .
```