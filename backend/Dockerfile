FROM golang:1.24-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /skillhub ./cmd/server

FROM alpine:3.21
WORKDIR /srv/skillhub
COPY --from=build /skillhub /usr/local/bin/skillhub
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/skillhub"]
