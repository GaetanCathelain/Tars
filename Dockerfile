FROM golang:1.23-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o tars ./cmd/tars

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /build/tars /usr/local/bin/tars
EXPOSE 8080
CMD ["tars"]
