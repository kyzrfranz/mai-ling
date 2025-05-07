FROM golang:1.24 AS build
RUN apt update
RUN apt install -y ca-certificates && update-ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make build-linux-amd64

FROM debian:bookworm-slim AS dockerize
RUN apt-get update && apt-get install -y \
    webp && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /app/build/mailer-amd64-linux /mailer
COPY --from=build /app/data /data

EXPOSE 8080
ENTRYPOINT ["/mailer"]



docker -e AUTH_KEY=raltUdabCyun -e MONGO_COLLECTION=live-bucket-1 -e MONGO_DB_NAME=buntesdach -e MONGO_URI=mongodb+srv://api-user:YovbjXMwqNPNnVhp@cluster0.uixtb.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0 run -t eu.gcr.io/buntesdach/mailer:latest
