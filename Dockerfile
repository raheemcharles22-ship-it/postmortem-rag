# stage 1: build
FROM golang:1.26.5 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

# stage 2: run
FROM alpine
COPY --from=build /server /server
ENTRYPOINT ["/server"]