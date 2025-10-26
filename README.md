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

### API Endpoints

#### User Management

- `POST /v1/users` - Create new user
- `GET /v1/users` - List users (with pagination)

#### Hotel Management

- `GET /v1/hotels/:hotel_id` - Get hotel price information

#### Favorites Management

- `POST /v1/users/:user_id/favorites` - Add hotel to favorites
- `GET /v1/users/:user_id/favorites` - List user favorites

#### System

- `GET /v1/healthcheck` - Health check endpoint

## Demo

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

### Kubernetes Deployment

1. **Create Kind cluster**

   ```bash
   make cluster
   ```

2. **Deploy application to Kubernetes**

   ```bash
   make app
   ```

3. **Check deployment status**

   ```bash
   kubectl get pods -n nuitee-challenge
   kubectl get services -n nuitee-challenge
   ```

4. **Clean up**
   ```bash
   make destroy
   ```

## Improvement Ideas

- **JWT-based Authentication** - Implement secure token-based authentication for user sessions
- **Rate Limiting** - Add API rate limiting to prevent abuse and ensure fair usage
- **Input Validation** - Strengthen input sanitization and validation middleware

- **Password Reset Flow** - Implement secure password reset with email verification tokens
- **Email Verification** - Add email confirmation for new user registrations
- **Profile Management** - Allow users to update their profile information
- **Account Deactivation** - Provide users with account deactivation/deletion options

- **ArgoCD Integration** - Implement GitOps workflow with ArgoCD for automated deployments
- **Rollback Capabilities** - Implement automated rollback mechanisms for failed deployments

- **Amazon Integration** - EKS for Kubernetes deployment, RDS for managed PostgreSQL, ElastiCache for Redis caching, Load Balancer for traffic distribution, S3 for logs and backups, Secrets Manager for credential management

- **Prometheus Metrics** - Add custom metrics for API performance, database queries, and business logic
- **Grafana Dashboards** - Create comprehensive dashboards for system health and business metrics
- **Alerting System** - Implement alerting for system failures, performance degradation, and price alerts
- **Distributed Tracing** - Add request tracing across services for better debugging
