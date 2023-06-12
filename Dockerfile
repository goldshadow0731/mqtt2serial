# Build stage
FROM golang:1.20.4 AS build-env
WORKDIR /app
COPY * /app
ARG GOARCH
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH:-amd64} go build -o mqtt2serial

# Final stage
FROM scratch
COPY --from=build-env /app/mqtt2serial /
CMD ["/mqtt2serial"]
