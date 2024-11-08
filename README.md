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
- `IDENFY_CALLBACK_SIGN_KEY`: Callback signing key for iDenfy webhooks (required) (note: should match the signing key in iDenfy dashboard for the related environment and should be at least 32 characters long)
- `IDENFY_WHITELISTED_IPS`: Comma-separated list of whitelisted IPs for iDenfy callbacks
- `IDENFY_DEV_MODE`: Enable development mode for iDenfy integration (default: false) (note: works only in iDenfy dev environment, enabling it in test or production environment will cause iDenfy to reject the requests)
- `IDENFY_CALLBACK_URL`: URL for iDenfy verification update callbacks. (example: `https://{KYC-SERVICE-DOMAIN}/webhooks/idenfy/verification-update`)
- `IDENFY_NAMESPACE`: Namespace for isolating diffrent TF KYC verifier services data in same iDenfy backend (default: "") (note: if you are using the same iDenfy backend for multiple services on same tfchain network, you can set this to the unique identifier of the service to isolate the data. don't touch unless you know what you are doing)

### TFChain Configuration

- `TFCHAIN_WS_PROVIDER_URL`: WebSocket provider URL for TFChain (default: "wss://tfchain.grid.tf")

### Verification Settings

- `VERIFICATION_SUSPICIOUS_VERIFICATION_OUTCOME`: Outcome for suspicious verifications (default: "APPROVED")
- `VERIFICATION_EXPIRED_DOCUMENT_OUTCOME`: Outcome for expired documents (default: "REJECTED")
- `VERIFICATION_MIN_BALANCE_TO_VERIFY_ACCOUNT`: Minimum balance in unitTFT required to verify an account (default: 10000000)
- `VERIFICATION_ALWAYS_VERIFIED_IDS`: Comma-separated list of TFChain SS58Addresses that are always verified (default: "")

### Rate Limiting

#### IP-based Rate Limiting

- `IP_LIMITER_MAX_TOKEN_REQUESTS`: Maximum number of token requests per IP (default: 4)
- `IP_LIMITER_TOKEN_EXPIRATION`: Token expiration time in minutes (default: 1440)

#### ID-based Rate Limiting

- `ID_LIMITER_MAX_TOKEN_REQUESTS`: Maximum number of token requests per ID (default: 4)
- `ID_LIMITER_TOKEN_EXPIRATION`: Token expiration time in minutes (default: 1440)

### Challenge Configuration

- `CHALLENGE_WINDOW`: Time window in seconds for challenge validation (default: 8)
- `CHALLENGE_DOMAIN`: Current service domain name for challenge validation (required) (example: `tfkyc.dev.grid.tf`)

### Logging

- `DEBUG`: Enable debug logging (default: false)

To configure these options, you can either set them as environment variables or include them in your `.env` file.

Regarding the iDenfy signing key, it's best to use key composed of alphanumeric characters to avoid such issues.
You can generate a random key using the following command:

```bash
cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1
```

Refer to `internal/configs/config.go` for the implementation details of these configuration options.

## Running the Application

### Using Docker Compose

First make sure to create and set the environment variables in the `.app.env`, `.db.env` files.
Examples can be found in `.app.env.example`, `.db.env.example`.
In beta releases, we include the mongo-express container, but you can opt to disable it.

To start only the core services (API and MongoDB) using Docker Compose:

```bash
docker compose up -d
```

To include mongo-express for development, make sure to create and set the environment variables in the `.express.env` file as well, then run:

```bash
docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d
```

To start only mongo-express if core services are already running, run:

```bash
docker compose -f docker-compose.dev.yml up -d mongo-express
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

- `GET /api/v1/status`
  - Get verification status
  - Query Parameters (at least one required):
    - `client_id`: TFChain SS58Address (48 chars)
    - `twin_id`: Twin ID
  - Responses:
    - `200`: Success
    - `400`: Bad request
    - `404`: Not found

### Webhook Endpoints

- `POST /webhooks/idenfy/verification-update`
  - Process verification update from iDenfy
  - Required Headers:
    - `Idenfy-Signature`: Verification signature
  - Responses:
    - `200`: Success
    - `400`: Bad request

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

### Miscellaneous

- `GET /api/v1/version`
  - Get application version
  - Responses:
    - `200`: Returns application version
      - `version`: Application version

- `GET /api/v1/configs`
  - Get application configurations
  - Responses:
    - `200`: Returns application configurations

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

### Creating database dump

Most of the normal tools will work, although their usage might be a little convoluted in some cases to ensure they have access to the mongod server. A simple way to ensure this is to use docker exec and run the tool from the same container, similar to the following:

```bash
docker exec <mongo_db_container_name> sh -c 'exec mongodump -d <database_name> --archive' > /some/path/on/your/host/all-collections.archive
```

## Production

Refer to the [Production Setup](./docs/production.md) documentation for production setup details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the Apache 2.0 License. See the `LICENSE` file for more details.
