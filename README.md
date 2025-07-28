# Phinex Blog API

This is the backend API for the Phinex Blog application, built with Go and Fiber.

## Table of Contents
- [Phinex Blog API](#phinex-blog-api)
  - [Table of Contents](#table-of-contents)
  - [Features](#features)
  - [Technologies Used](#technologies-used)
  - [Setup and Installation](#setup-and-installation)
    - [Prerequisites](#prerequisites)
    - [Environment Variables](#environment-variables)
    - [Running with Docker Compose](#running-with-docker-compose)
    - [Running Locally](#running-locally)
  - [API Documentation](#api-documentation)
  - [Running Tests](#running-tests)

## Features
- User Authentication (Registration, Login)
- Blog Post Management (CRUD)
- Commenting System
- Likes and Shares
- User Following

## Technologies Used
- **Go**: Primary programming language
- **Fiber**: Fast, Express-inspired web framework
- **GORM**: ORM library for Go
- **PostgreSQL**: Database
- **Docker**: Containerization
- **Caddy**: Reverse proxy and web server
- **Swagger**: API Documentation

## Setup and Installation

### Prerequisites
- Go (version 1.18 or higher)
- Docker and Docker Compose
- PostgreSQL (if running locally without Docker Compose)

### Environment Variables
Create a `.env` file in the root directory of the project with the following variables:

```
PORT=5000

POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=your_username
POSTGRES_PASSWORD=your_password
POSTGRES_DB=your_database_name

JWT_SECRET=your_jwt_secret_key
```

### Running with Docker Compose

1.  **Build and run the containers:**
    ```bash
    docker-compose up --build
    ```
    This will build the Go application image, set up the PostgreSQL database, and start the Caddy reverse proxy.

2.  The API will be accessible via Caddy, typically on `http://localhost` or `https://phinex.blog.epsierra.com` if you configure your hosts file and Caddyfile for production.

### Running Locally

1.  **Install Go dependencies:**
    ```bash
    go mod tidy
    ```

2.  **Ensure PostgreSQL is running** and your `.env` file is configured correctly for local PostgreSQL.

3.  **Run the application:**
    ```bash
    go run main.go
    ```
    The API will be available at `http://localhost:5000` (or the port specified in your `.env` file).

## API Documentation

Swagger UI is integrated for API documentation. Once the application is running, you can access it at:

-   **Local Development**: `http://localhost:5000/swagger-docs/index.html` (or your configured port)
-   **Docker Compose (via Caddy)**: `http://localhost/swagger-docs/index.html` (or your configured domain)

## Running Tests

To run the end-to-end API tests, navigate to the project root and execute:

```bash
go test ./src/test
```
