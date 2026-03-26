[![Go Report Card](https://goreportcard.com/badge/github.com/unshade/citadelle)](https://goreportcard.com/report/github.com/unshade/citadelle)
[![License](https://img.shields.io/github/license/unshade/citadelle)](LICENSE)
[![codecov](https://codecov.io/github/unshade/citadelle/branch/main/graph/badge.svg?token=YG6U49FPUL)](https://codecov.io/github/unshade/citadelle)
[![Release](https://img.shields.io/github/release/unshade/citadelle.svg)](https://github.com/unshade/citadelle/releases)

# Citadelle

Citadelle is the backend for an end-to-end encrypted file storage system. The server stores only encrypted blobs. Keys, file content, and paths are all encrypted client-side before being sent. The server never has access to plaintext data.

Built with Go, [Fuego](https://github.com/go-fuego/fuego), and PostgreSQL.

## Prerequisites

- Go 1.24+
- Docker (for the local PostgreSQL instance)

## Dev Setup

**1. Clone and install dependencies**

```bash
git clone https://github.com/unshade/citadelle.git
cd citadelle
go mod download
```

**2. Configure environment**

```bash
cp .env.example .env
```

Edit `.env` if needed. The defaults work out of the box with the dev Docker setup.

**3. Start the database**

```bash
docker compose -f docker-compose.dev.yml up -d
```

> The Docker Compose file matches the credentials in `.env.example`. If you change them, update both files.

**4. Run the server**

```bash
go run main.go server
```

The server starts on `http://localhost:8080`. Database migrations run automatically on startup.

## API

The OpenAPI spec is available at `doc/openapi.json`. When the server is running, Fuego also serves it at `http://localhost:8080/swagger`.
