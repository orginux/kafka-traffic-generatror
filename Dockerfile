FROM golang:1.20 AS build-stage
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /kafka-traffic-generator

FROM gcr.io/distroless/base-debian11:latest AS build-release-stage
WORKDIR /
COPY --from=build-stage /kafka-traffic-generator /kafka-traffic-generator
USER nonroot:nonroot
ENTRYPOINT ["/kafka-traffic-generator"]
