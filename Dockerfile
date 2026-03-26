FROM golang:1.26-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -trimpath \
    -o /citadelle \
    .

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /citadelle /citadelle

EXPOSE 8080

ENTRYPOINT ["/citadelle", "server"]
