FROM golang:1.20 AS build-stage
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /kafka-traffic-generator

FROM gcr.io/distroless/base-debian11:sha256-bff68ceffd34b9cf28686da8b11c2ab23c1220c785d2c7f3d319eeda8aeb5035.sig AS build-release-stage
WORKDIR /
COPY --from=build-stage /kafka-traffic-generator /kafka-traffic-generator
USER nonroot:nonroot
ENTRYPOINT ["/kafka-traffic-generator"]
