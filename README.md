# go-microservices

This is a test project to practice go and microservices.

## The project consists of the following main components:

- Account Service
- Catalog Service
- Order Service
- GraphQL API Gateway

## Each service has its own database:

- Account and Order services use PostgreSQL
- Catalog service uses Elasticsearch

## Startup

`docker compose up -d --build`
