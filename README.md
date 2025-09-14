# 2do

**An over-engineered todo app**

---

## Overview

**2do** is a todo list application built with Go, designed with a more elaborate architecture than typically necessary—hence the "over-engineered" moniker. It uses MySQL as its database backend and serves as a showcase for Go back-end development, Docker containerization, and database migrations.

---

## Features

- RESTful API for managing todo items
- Docker support for easy setup with `Dockerfile` and `docker-compose.yml`
- Database migrations included for schema versioning

---

## Prerequisites

- [Go](https://golang.org/) (recent stable version)
- [Docker](https://www.docker.com/) and Docker Compose (optional but recommended)

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

* Builds the app and any required services (e.g., database).
* Access the API at `http://localhost:8080/`.

**Note (Optional):** The API port is configurable. Set PORT in .env to control the binding inside the container, and adjust the ports mapping in docker-compose.yml to decide how it’s exposed on your host machine.

---

## Usage

Interact with the todo API endpoints (e.g., via `curl` or Postman):

* `GET /api/v0/todos` — list all todos
* `POST /api/v0/todos` — create a new todo
* `GET /api/v0/todos/{id}` — retrieve a specific todo
* `PUT /api/v0/todos/{id}` — update a todo
* `DELETE /api/v0/todos/{id}` — delete a todo
