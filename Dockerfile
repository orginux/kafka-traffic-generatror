FROM golang:1.21.2-bookworm AS build-stage
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /kafka-traffic-generator

FROM scratch
WORKDIR /
COPY --from=build-stage /kafka-traffic-generator /kafka-traffic-generator
ENTRYPOINT ["/kafka-traffic-generator"]
