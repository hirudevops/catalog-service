# Build stage
FROM golang:1.24.12-alpine AS build
WORKDIR /src

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/catalog-service ./cmd/server

# Run stage
FROM alpine:3.20
WORKDIR /

RUN apk add --no-cache \
	ca-certificates \
	tzdata \
	curl \
	busybox-extras \
	vim \
	bind-tools \
	netcat-openbsd

RUN addgroup -g 65532 nonroot \
	&& adduser -D -H -u 65532 -G nonroot nonroot

COPY --from=build /out/catalog-service /catalog-service
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo

USER nonroot
EXPOSE 8080
ENTRYPOINT ["/catalog-service"]