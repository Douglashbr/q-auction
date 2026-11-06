FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/user-service ./cmd/user

# alpine (e não scratch) para ter o wget do healthcheck.
FROM alpine:3.22
RUN adduser -D -H app
COPY --from=build /out/user-service /usr/local/bin/user-service
USER app
EXPOSE 3000
CMD ["user-service"]
