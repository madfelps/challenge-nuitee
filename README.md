# Hotel Price Monitor Project

This project aims to validate a hotel price monitor project that tracks hotel prices and alerts users when their target prices are reached.

Throughout this documentation, all the technical aspects (data design and tools) are discussed, including their tradeoffs.

## General Overview

The Hotel Price Monitor is a RESTful API service that allows users to:

- Register and manage user accounts
- Add hotels to their favorites list with target prices
- Monitor hotel prices in real-time using the LiteAPI
- Receive price alerts when current prices drop below target prices

The system consists of a Go-based API server, PostgreSQL database, and a background price monitoring routine that checks prices every minute.

## Stack Setup

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- Kind (Kubernetes in Docker)
- Terraform
- Make
- kubectl

### Technology Stack & Decisions

#### Backend

- **Go 1.21+** - High performance, excellent concurrency support for price monitoring, and strong ecosystem for web APIs
- **PostgreSQL 16** - ACID guarantees for user data
- **DBeaver** - Local development and facilitate schema visualization
- **Docker & Docker Compose** - Development environment using containers

#### Infrastructure

- **Kubernetes (Kind)** - Local cluster testing without cloud dependencies
- **Terraform** - Infrastructure as Code for reproducible deployments and infrastructure consistency
- **Make** - Script build automation

## Schema Design

The database schema consists of two tables that exposes price monitoring registers and user management. The background job will analyse the whole **users_favorites** table, and call LiteAPI to check what is the current hotel price. If the price is below the target price, the application alerts (using log) this event.

Concerning the **users** table, a user password is stored in format salt:hashPassword. This decision was made to prevent rainbow table attacks.

![Database Schema Diagram](docs/tables.jpg?raw=true "Database schema design")

## CI Pipeline

A simple CI pipeline was build using Github Actions with the following jobs:

- Git leaks: In order to check for leaked secrets, such as API KEYS
- Test: To run our unit tests (and integration tests in the future)
- Build: To package our application in a Docker image
- Deploy: To push the generated image to Dockerhub

![CI Pipeline workflow](https://raw.githubusercontent.com/madfelps/challenge-nuitee/main/docs/pipeline.jpg "Pipeline workflow")

## API Endpoints

### User Management

- `POST /v1/users` - Create new user
- `GET /v1/users` - List users (with pagination)

### Hotel Management

- `GET /v1/hotels/:hotel_id` - Get hotel price information

### Favorites Management

- `POST /v1/users/:user_id/favorites` - Add hotel to favorites
- `GET /v1/users/:user_id/favorites` - List user favorites

### System

- `GET /v1/healthcheck` - Health check endpoint

## Demo

![nuitee](https://github.com/user-attachments/assets/64290b85-7e54-4442-8fff-5a3630055805)

## How To Run

### Local Development with Docker

1. **Start the application**

   ```bash
   make run
   ```

2. **Stop the application**
   ```bash
   make down
   ```

The API will be available at `http://localhost:4000`

### Environment Variables Configuration

Before running the application, you need to configure the environment variables. Create a `.env` file in the project root with the following values:

```bash
DATABASE_USER=nuitee
DATABASE_PASSWORD=1234
DATABASE_NAME=nuitee
DATABASE_DSN=postgresql://nuitee:1234@db:5432/nuitee?sslmode=disable

LITE_API_KEY=your_lite_api_key_here
```

**Getting your LiteAPI Key:**

1. Visit [LiteAPI Dashboard](https://dashboard.liteapi.travel/apikeys)
2. Log in to your personal account
3. Copy your API key and replace `your_lite_api_key_here` in the `.env` file. Personally, I suggest you to use an API KEY for sandbox environment to run this project.

## Improvement Ideas

- **JWT-based Authentication** - Implement secure token-based authentication for user sessions
- **Memory-cache usage** - Implement cache in memory (such as Redis) to improve application performance

- **Password Reset Flow** - Implement secure password reset with email verification tokens
- **Profile Management** - Allow users to update their profile information

- **ArgoCD Integration** - Implement GitOps workflow with ArgoCD for automated deployments and rollbacks

- **Amazon Integration** - EKS for Kubernetes deployment, RDS for managed PostgreSQL, ElastiCache for Redis caching, Load Balancer for traffic distribution, S3 for logs and backups and Secrets Manager for credential management

- **Observability** - Add custom metrics for API performance using Prometheus, create Grafana dashboards and implement alerts for system failures
