# Kubernetes Deployment with Kind

This directory contains Terraform configuration and Makefile commands to deploy the Nuitee Challenge application to a local Kubernetes cluster using Kind.

## Quick Start

### 1. Create Kind Cluster

```bash
make cluster
```

### 2. Deploy Application

```bash
make app
```

### 3. Check Status

```bash
make status
```

### 4. Access API

```bash
make port-forward
# API will be available at http://localhost:4000
```

## Available Commands

### 🚀 Cluster Management

- `make cluster` - Create Kind cluster
- `make destroy-cluster` - Destroy Kind cluster

### 📱 Application Management

- `make app` - Deploy application to cluster
- `make destroy-app` - Destroy application (keep cluster)

### 🔍 Monitoring & Debugging

- `make status` - Check cluster and application status
- `make port-forward` - Port forward API to localhost:4000
- `make logs-api` - View API logs
- `make logs-postgres` - View PostgreSQL logs

### 🌍 Environment Management

- `make dev` - Deploy to development
- `make prod` - Deploy to production
- `make clean` - Clean Terraform state

## Architecture

The deployment includes:

- **Kind Cluster**: Local Kubernetes cluster
- **Namespace**: `nuitee-challenge`
- **API Deployment**: Go application with LiteAPI integration
- **PostgreSQL**: Database with persistent storage
- **Services**: Internal communication between components
- **ConfigMaps & Secrets**: Configuration and sensitive data

## Configuration

- **Cluster Config**: `kind-config.yaml`
- **Terraform Variables**: `terraform.tfvars`
- **Environment Files**: `environments/*.tfvars`

## Troubleshooting

### Check Cluster Status

```bash
kubectl cluster-info
kubectl get nodes
```

### Check Application Status

```bash
kubectl get pods -n nuitee-challenge
kubectl get services -n nuitee-challenge
```

### View Logs

```bash
make logs-api
make logs-postgres
```

### Port Forward

```bash
make port-forward
# Then test: curl http://localhost:4000/v1/healthcheck
```

## Clean Up

### Destroy Application Only

```bash
make destroy-app
```

### Destroy Everything

```bash
make destroy-cluster
```
