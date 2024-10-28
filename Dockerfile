FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o tfgrid-kyc cmd/api/main.go

FROM alpine:3.19

COPY --from=builder /app/tfgrid-kyc .
RUN apk --no-cache add curl

ENTRYPOINT ["/tfgrid-kyc"]

EXPOSE 8080
