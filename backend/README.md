# Backend

Backend comprises many different microservices. It uses go workspace to manage dependencies. Shared functionality is located in the `common` directory.

## Docker

Dockerizing microservice on example of `logger`.

```bash
docker build -t $REGISTRY_URL/alerting-platform-logger:latest -f services/logger/Dockerfile .

docker push $REGISTRY_URL/alerting-platform-logger:latest
```

`REGISTRY_URL` value can be found using `terraform output`.
