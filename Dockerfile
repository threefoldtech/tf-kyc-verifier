FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN VERSION=$(git describe --tags --always) && \
    CGO_ENABLED=0 GOOS=linux go build -o tfkycv -ldflags "-X github.com/threefoldtech/tf-kyc-verifier/internal/build.Version=$VERSION" cmd/api/main.go

FROM alpine:3.19

COPY --from=builder /app/tfkycv .
RUN apk --no-cache add curl

ENTRYPOINT ["/tfkycv"]

EXPOSE 8080
