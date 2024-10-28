# TF KYC Service

## Overview

TF KYC Service is a Go-based service that provides Know Your Customer (KYC) functionality for the TF Grid. It integrates with iDenfy for identity verification.

## Features

- Identity verification using iDenfy
- Blockchain integration with TFChain (Substrate-based)
- MongoDB for data persistence
- RESTful API endpoints for KYC operations
- Swagger documentation
- Containerized deployment

## Prerequisites

- Go 1.22+
- MongoDB 4.4+
- Docker and Docker Compose (for containerized deployment)
- iDenfy API credentials

## Installation

1. Clone the repository:

    ```bash
    git clone https://github.com/yourusername/tf-kyc-verifier.git
    cd tf-kyc-verifier
    ```

2. Set up your environment variables:

    ```bash
    cp .app.env.example .app.env
    cp .db.env.example .db.env
    ```

Edit `.app.env` and `.db.env` with your specific configuration details.

## Configuration

The application uses environment variables for configuration. Here's a list of all available configuration options:

### Database Configuration

- `MONGO_URI`: MongoDB connection URI (default: "mongodb://localhost:27017")
- `DATABASE_NAME`: Name of the MongoDB database (default: "tf-kyc-db")

### Server Configuration

- `PORT`: Port on which the server will run (default: "8080")

### iDenfy Configuration

- `IDENFY_API_KEY`: API key for iDenfy service (required) (note: make sure to use correct iDenfy API key for the environment dev, test, and production) (iDenfy dev -> TFChain Devnet, iDenfy test -> TFChain QAnet, iDenfy prod -> TFChain Testnet and Mainnet)
- `IDENFY_API_SECRET`: API secret for iDenfy service (required)
- `IDENFY_BASE_URL`: Base URL for iDenfy API (default: "<https://ivs.idenfy.com>")
- `IDENFY_CALLBACK_SIGN_KEY`: Callback signing key for iDenfy webhooks (required) (note: should match the signing key in iDenfy dashboard for the related environment)
- `IDENFY_WHITELISTED_IPS`: Comma-separated list of whitelisted IPs for iDenfy callbacks
- `IDENFY_DEV_MODE`: Enable development mode for iDenfy integration (default: false) (note: works only in iDenfy dev environment, enabling it in test or production environment will cause iDenfy to reject the requests)
- `IDENFY_CALLBACK_URL`: URL for iDenfy verification update callbacks. (example: `https://{KYC-SERVICE-DOMAIN}/webhooks/idenfy/verification-update`)

### TFChain Configuration

- `TFCHAIN_WS_PROVIDER_URL`: WebSocket provider URL for TFChain (default: "wss://tfchain.grid.tf")

### Verification Settings

- `VERIFICATION_SUSPICIOUS_VERIFICATION_OUTCOME`: Outcome for suspicious verifications (default: "verified")
- `VERIFICATION_EXPIRED_DOCUMENT_OUTCOME`: Outcome for expired documents (default: "unverified")
- `VERIFICATION_MIN_BALANCE_TO_VERIFY_ACCOUNT`: Minimum balance required to verify an account (default: 10000000)

### Rate Limiting

#### IP-based Rate Limiting

- `IP_LIMITER_MAX_TOKEN_REQUESTS`: Maximum number of token requests per IP (default: 4)
- `IP_LIMITER_TOKEN_EXPIRATION`: Token expiration time in hours (default: 24)

#### ID-based Rate Limiting

- `ID_LIMITER_MAX_TOKEN_REQUESTS`: Maximum number of token requests per ID (default: 4)
- `ID_LIMITER_TOKEN_EXPIRATION`: Token expiration time in hours (default: 24)

### Challenge Configuration

- `CHALLENGE_WINDOW`: Time window in seconds for challenge validation (default: 8)
- `CHALLENGE_DOMAIN`: Current service domain name for challenge validation (required) (example: `tfkyc.dev.grid.tf`)

### Logging

- `DEBUG`: Enable debug logging (default: false)

To configure these options, you can either set them as environment variables or include them in your `.env` file.

Refer to `internal/configs/config.go` for the implementation details of these configuration options.

## Running the Application

### Using Docker Compose

To start the server and MongoDB using Docker Compose:

```bash
docker-compose up -d --build
```

### Running Locally

To run the application locally:

1. Ensure MongoDB is running and accessible.
2. export the environment variables:

    ```bash
    set -a
    source .app.env
    set +a
    ```

3. Run the application:

    ```bash
    go run cmd/api/main.go
    ```

## API Endpoints

### Client Endpoints

#### Token Management

- `POST /api/v1/token`
  - Get or create a verification token
  - Required Headers:
    - `X-Client-ID`: TFChain SS58Address (48 chars)
    - `X-Challenge`: Hex-encoded message `{api-domain}:{timestamp}`
    - `X-Signature`: Hex-encoded sr25519|ed25519 signature (128 chars)
  - Responses:
    - `200`: Existing token retrieved
    - `201`: New token created
    - `400`: Bad request
    - `401`: Unauthorized
    - `402`: Payment required
    - `409`: Conflict
    - `500`: Internal server error

#### Verification

- `GET /api/v1/data`
  - Get verification data for a client
  - Required Headers:
    - `X-Client-ID`: TFChain SS58Address (48 chars)
    - `X-Challenge`: Hex-encoded message `{api-domain}:{timestamp}`
    - `X-Signature`: Hex-encoded sr25519|ed25519 signature (128 chars)
  - Responses:
    - `200`: Success
    - `400`: Bad request
    - `401`: Unauthorized
    - `404`: Not found
    - `500`: Internal server error

- `GET /api/v1/status`
  - Get verification status
  - Query Parameters (at least one required):
    - `client_id`: TFChain SS58Address (48 chars)
    - `twin_id`: Twin ID
  - Responses:
    - `200`: Success
    - `400`: Bad request
    - `404`: Not found
    - `500`: Internal server error

### Webhook Endpoints

- `POST /webhooks/idenfy/verification-update`
  - Process verification update from iDenfy
  - Required Headers:
    - `Idenfy-Signature`: Verification signature
  - Responses:
    - `200`: Success
    - `400`: Bad request
    - `500`: Internal server error

- `POST /webhooks/idenfy/id-expiration`
  - Process document expiration notification (Not implemented)
  - Responses:
    - `501`: Not implemented

### Health Check

- `GET /api/v1/health`
  - Check service health status
  - Responses:
    - `200`: Returns health status
      - `healthy`: All systems operational
      - `degraded`: Some systems experiencing issues

### Documentation

- `GET /docs`
  - Swagger documentation interface
  - Provides interactive API documentation and testing interface

Refer to the Swagger documentation at `/docs` endpoint for detailed information about request/response formats and examples.

## Swagger Documentation

Swagger documentation is available. To view it, run the application and navigate to the `/docs` endpoint in your browser.

## Project Structure

- `cmd/`: Application entrypoints
  - `api/`: Main API server
- `internal/`: Internal packages
  - `clients/`: External service clients
  - `configs/`: Configuration handling
  - `errors/`: Custom error types
  - `handlers/`: HTTP request handlers
  - `logger/`: Logging configuration
  - `middlewares/`: HTTP middlewares
  - `models/`: Data models
  - `repositories/`: Data access layer
  - `responses/`: API response structures
  - `server/`: Server setup and routing
  - `services/`: Business logic
- `api/`: API documentation
  - `docs/`: Swagger documentation files
- `.github/`: GitHub specific files
  - `workflows/`: GitHub Actions workflows
- `scripts/`: Utility and Development scripts
- `docs/`: Documentation

- Configuration files:
  - `.app.env.example`: Example application environment variables
  - `.db.env.example`: Example database environment variables
  - `Dockerfile`: Container build instructions
  - `docker-compose.yml`: Multi-container Docker setup
  - `go.mod`: Go module definition
  - `go.sum`: Go module checksums

## Development

### Running Tests

To run the test suite:

TODO: Add tests

### Building the Docker Image

To build the Docker image:

```bash
docker build -t tf_kyc_verifier .
```

### Running the Docker Container

To run the Docker container and use .env variables:

```bash
docker run -d -p 8080:8080 --env-file .app.env tf_kyc_verifier
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the Apache 2.0 License. See the `LICENSE` file for more details.
