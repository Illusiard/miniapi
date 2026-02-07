FROM golang:1.25.7-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /miniapi ./cmd/server

FROM alpine:3.22
WORKDIR /app
COPY --from=build /miniapi /app/miniapi
ENTRYPOINT ["/app/miniapi"]
