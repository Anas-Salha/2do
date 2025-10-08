# 2do

**An over-engineered todo app backend**

---

## Prerequisites

- [Go](https://golang.org/)
- [Docker](https://www.docker.com/) and Docker Compose

---

## Getting Started
1. Copy the example environment file and adjust values as needed:
```bash
cp .env.example .env
```

2. Start the app with Docker
```bash
docker compose up --build
```

3. Use curl (in terminal) or Swagger UI to try out API calls

* Access the API at `http://localhost:8080/`
* Access Swagger UI at `http://localhost:8081/`


**Note (Optional):** The API port is configurable. Set PORT in .env to control the binding inside the container, and adjust the ports mapping in docker-compose.yml to decide how it’s exposed on your host machine.

---

## REST API
**Documentation:** https://anas-salha.github.io/2do/

- OpenAPI schema: [`openapi.yaml`](./docs/openapi.yaml)
