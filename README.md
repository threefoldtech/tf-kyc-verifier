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
    cp .express.env.example .express.env # only if you are using mongo-express for development
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

- `IDENFY_API_KEY`: API key for iDenfy service (required) (Ensure the correct iDenfy API key is used for the respective environment: iDenfy dev for TFChain Devnet, iDenfy test for TFChain QAnet, iDenfy prod for TFChain Testnet and Mainnet.)
- `IDENFY_API_SECRET`: API secret for iDenfy service (required)
- `IDENFY_BASE_URL`: Base URL for iDenfy API (default: "<https://ivs.idenfy.com>")
- `IDENFY_CALLBACK_SIGN_KEY`: Callback signing key for iDenfy webhooks (required) (Must match the signing key configured in the iDenfy dashboard for the related environment and should be at least 32 characters long.)
- `IDENFY_WHITELISTED_IPS`: Comma-separated list of whitelisted IPs for iDenfy callbacks
- `IDENFY_DEV_MODE`: Enable development mode for iDenfy integration. When enabled, retrieving a verification token will simulate the iDenfy KYC flow, and the KYC service will receive verification update callbacks without actual iDenfy processing. (default: false) (Note: This mode is intended for iDenfy development environments only. Enabling it in test or production environments will cause iDenfy to reject requests.)
- `IDENFY_CALLBACK_URL`: URL for iDenfy verification update callbacks. (example: `https://{KYC-SERVICE-DOMAIN}/webhooks/idenfy/verification-update`)
- `IDENFY_NAMESPACE`: A namespace for isolating different TF KYC verifier services' data within the same iDenfy backend. (default: "") (Use this if you are running multiple KYC services on the same TFChain network and sharing an iDenfy backend, to ensure data isolation. Don't touch unless you know what you are doing!)

### TFChain Configuration

- `TFCHAIN_WS_PROVIDER_URL`: WebSocket provider URL for TFChain (default: "wss://tfchain.grid.tf" - This is typically for Mainnet. Adjust for Devnet, Testnet, or QAnet as needed.)

### Verification Settings

- `VERIFICATION_SUSPICIOUS_VERIFICATION_OUTCOME`: Outcome for suspicious verifications (default: "APPROVED")
- `VERIFICATION_EXPIRED_DOCUMENT_OUTCOME`: Outcome for expired documents (default: "REJECTED")
- `VERIFICATION_MIN_BALANCE_TO_VERIFY_ACCOUNT`: Minimum balance in uTFT required to verify an account (1 TFT = 10,000,000 uTFT) (default: 10000000) (Note: Can be set to 0 to disable this check, but be aware that this can lead to abuse)
- `VERIFICATION_ALWAYS_VERIFIED_IDS`: Comma-separated list of TFChain SS58Addresses that are always verified (default: "")
- `VERIFICATION_ALWAYS_VERIFIED_IDS_ONLY`: When this is true, the creation of KYC tokens is disabled on this network (default: false)

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

For implementation details, refer to `internal/configs/config.go`.

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
docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d mongo-express
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

### Authentication

The following endpoints require authentication using a challenge-response mechanism with the specified headers:

### Authentication Headers

When authentication is required, include these headers in your request:

| Header | Description | Required |
|--------|-------------|----------|
| `X-Client-ID` | The TFChain SS58Address of the client | Yes |
| `X-Challenge` | Hex-encoded message `{api-domain}:{timestamp}` | Yes |
| `X-Signature` | Hex-encoded sr25519 or ed25519 signature of the challenge | Yes |

For Sponsee authentication, include these headers in your request as additional required headers:

| Header | Description | Required |
|--------|-------------|----------|
| `X-Sponsee-ID` | The TFChain SS58Address of the sponsee | Yes |
| `X-Sponsee-Challenge` | Hex-encoded message `{api-domain}:{timestamp}` | Yes |
| `X-Sponsee-Signature` | Hex-encoded sr25519 or ed25519 signature of the sponsee | Yes |

### Endpoint Authentication Requirements

| Endpoint | Method | Authentication Required | Notes |
|----------|--------|-------------------------|-------|
| `/api/v1/token` | POST | Yes | Requires standard authentication |
| `/api/v1/data` | GET | Yes | Requires standard authentication |
| `/api/v1/sponsorships` | POST | Yes | Requires both sponsor and sponsee authentication |

### Sponsorships

#### Create Sponsorship

- `POST /api/v1/sponsorships`
  - Creates a new sponsorship between a KYC-verified sponsor and a sponsee
  - Required Headers:
    - `X-Client-ID`: Sponsor's TFChain SS58Address
    - `X-Sponsee-ID`: Sponsee's TFChain SS58Address
    - `X-Challenge`: Hex-encoded message `{api-domain}:{timestamp}` for sponsor
    - `X-Sponsee-Challenge`: Hex-encoded message `{api-domain}:{timestamp}` for sponsee
    - `X-Signature`: Sponsor's signature
    - `X-Sponsee-Signature`: Sponsee's signature
  - Responses:
    - `201`: Sponsorship created successfully
    - `400`: Bad request
    - `401`: Unauthorized
    - `403`: Forbidden
    - `404`: Not found
    - `409`: Conflict (sponsorship already exists)

#### List Sponsorships

- `GET /api/v1/sponsorships`
  - List sponsorships with optional filtering
  - Query Parameters (only one filter can be used at a time):
    - `sponsor_twin_id`: Filter by sponsor twin ID
    - `sponsee_twin_id`: Filter by sponsee twin ID
    - `sponsor_client_id`: Filter by sponsor client ID
    - `sponsee_client_id`: Filter by sponsee client ID
    - `limit`: Maximum results (default: 50, max: 100)
    - `offset`: Pagination offset (default: 0)
  - Responses:
    - `200`: Success (returns paginated list)
    - `400`: Bad request

### Verification

#### Get Verification Token

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
    - `403`: Forbidden
    - `409`: Conflict

#### Get Verification Data

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

#### Get Verification Status

- `GET /api/v1/status`
  - Get verification status
  - Query Parameters (at least one required):
    - `client_id`: TFChain SS58Address (48 chars)
    - `twin_id`: Twin ID
  - Responses:
    - `200`: Success
    - `400`: Bad request
    - `404`: Not found

### Service Information

#### Health Check

- `GET /api/v1/health`
  - Check service health status
  - Responses:
    - `200`: Service is healthy
      - `healthy`: All systems operational
      - `degraded`: Some systems experiencing issues
    - `503`: Service unavailable

#### Get Service Configs

- `GET /api/v1/configs`
  - Get current service configuration (sensitive values redacted)
  - Responses:
    - `200`: Returns application configurations

#### Get Service Version

- `GET /api/v1/version`
  - Get service version information
  - Responses:
    - `200`: Returns application version
      - `version`: Application version

#### API Documentation

- `GET /docs`
  - Swagger documentation interface
  - Provides interactive API documentation and testing interface

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

Refer to the Swagger documentation at `/docs` endpoint for detailed information about request/response formats and examples.

## Development

### Local Development with ngrok

For local development, you can use ngrok to receive iDenfy webhook callbacks. Here's how to set it up:

1. Install ngrok:

   ```bash
   # For Ubuntu/Debian
   curl -sSL https://ngrok-agent.s3.amazonaws.com/ngrok.asc \
   | sudo tee /etc/apt/trusted.gpg.d/ngrok.asc >/dev/null \
   && echo "deb https://ngrok-agent.s3.amazonaws.com buster main" \
   | sudo tee /etc/apt/sources.list.d/ngrok.list \
   && sudo apt update \
   && sudo apt install ngrok
   
   # For macOS (using Homebrew)
   brew install ngrok
   ```

2. Sign up for a [free account](https://ngrok.com/signup?ref=downloads) then:

   ```bash
   ngrok config add-authtoken <token>
   ```

3. Start ngrok (replace 8080 with your app's port if different):

   ```bash
   ngrok http http://localhost:8080
   ```

4. Update your `.app.env` with the ngrok URL. It is crucial to set `CHALLENGE_DOMAIN` to your ngrok URL and `IDENFY_CALLBACK_URL` to the full webhook endpoint, as iDenfy requires a publicly accessible URL for callbacks:

   ```env
   CHALLENGE_DOMAIN=your-ngrok-url.ngrok.io
   IDENFY_CALLBACK_URL=https://your-ngrok-url.ngrok.io/webhooks/idenfy/verification-update
   ```

5. Restart your application to apply the changes.

6. Use the ngrok URL to test the API and receive webhook callbacks from iDenfy.

**Note:** If you prefer an alternative to ngrok, you can use `localtunnel`:

```bash
# Alternative using localtunnel
npx localtunnel --port 8080 --subdomain your-subdomain
```

### Project Structure

```text
.
├── api/                    # API documentation (Swagger)
├── cmd/                    # Main application entry points
│   └── api/                # API server entry point
├── configs/                # Configuration files
├── deployments/            # Deployment configurations
├── docs/                   # Documentation
├── internal/               # Private application code
│   ├── clients/            # External service clients
│   ├── config/             # Configuration handling
│   ├── handlers/           # HTTP request handlers
│   ├── middleware/         # HTTP middleware
│   ├── models/             # Data models
│   ├── repository/         # Data access layer
│   ├── responses/          # Response formatters
│   └── services/           # Business logic
├── scripts/                # Utility scripts
├── .app.env.example        # Example environment variables
├── .db.env.example         # Example database environment variables
├── docker-compose.yml      # Docker Compose configuration
├── go.mod                  # Go module definition
```

## Testing

Run unit tests:

```bash
go test ./...
```

Run tests with coverage:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

Run specific test:

```bash
go test -run TestFunctionName
```

## Building the Docker Image

To build the Docker image:

```bash
docker build -t tf_kyc_verifier .
```

## Running the Docker Container

To run the Docker container and use .env variables:

```bash
docker run -d -p 8080:8080 --env-file .app.env tf_kyc_verifier
```

## Creating database dump

Most of the normal tools will work, although their usage might be a little convoluted in some cases to ensure they have access to the mongod server. A simple way to ensure this is to use docker exec and run the tool from the same container, similar to the following:

```bash
#!/bin/bash
# mongo_backup.sh

# Install Environment file
source .db.env

docker exec tf_kyc_db mongodump --username $MONGO_INITDB_ROOT_USERNAME  --password $MONGO_INITDB_ROOT_PASSWORD  --authenticationDatabase admin --db tfgrid-kyc-db --archive=mongo.kyc.archive.dump
docker cp tf_kyc_db:/mongo.kyc.archive.dump mongo.kyc.archive.dump
```

### Restoring database dump

To restore the previously created backup, you can use a similar script as follows:

```bash
#!/bin/bash
# mongo_restore.sh

# Install Environment file
source .db.env

docker cp mongo.kyc.archive.dump tf_kyc_db:/mongo.kyc.archive.dump
docker exec tf_kyc_db mongorestore --username $MONGO_INITDB_ROOT_USERNAME  --password $MONGO_INITDB_ROOT_PASSWORD  --authenticationDatabase admin --nsInclude='tfgrid-kyc-db.*' --archive=mongo.kyc.archive.dump
```

## Production

Refer to the [Production Setup](./docs/production.md) documentation for production setup details.

## Contributing

### Development Workflow

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature-name`
3. Make your changes
4. Run tests: `make test`
5. Run linter: `make lint`
6. Format your code: `make fmt`
7. Generate swagger docs when needed: `make swagger`
8. Commit your changes: `git commit -m "Add your feature"`
9. Push to the branch: `git push origin feature/your-feature-name`
10. Open a pull request

### Makefile Commands

The project includes a `Makefile` with several useful commands for development and maintenance:

- `make test`: Run all unit tests.
- `make lint`: Run the linter to check code style and quality.
- `make fmt`: Format the Go source code according to standard Go formatting.
- `make swagger`: Generate or update Swagger API documentation.
- `make help`: Display a list of all available `make` commands and their descriptions.

### Code Style

- Follow standard Go formatting
- Write tests for new functionality
- Document public functions and types
- Keep commits focused and atomic

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the Apache 2.0 License. See the `LICENSE` file for more details.
