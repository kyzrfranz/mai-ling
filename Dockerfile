# Use --platform=$BUILDPLATFORM to run the Go compiler natively on your Mac
FROM --platform=$BUILDPLATFORM golang:1.26 AS build
RUN apt update && apt install -y ca-certificates && update-ca-certificates
WORKDIR /app

# Step 1: Copy and download dependencies (cached)
COPY go.mod go.sum ./
RUN go mod download

# Step 2: Copy the rest of the source
COPY . .

# Step 3: Cross-compile for the target architecture
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /mailer ./cmd/server.go

# Final Stage: The runtime image
FROM debian:bookworm-slim AS dockerize
RUN apt-get update && apt-get install -y \
    webp \
    ca-certificates && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /mailer /mailer
COPY --from=build /app/data /data

EXPOSE 8080
ENTRYPOINT ["/mailer"]