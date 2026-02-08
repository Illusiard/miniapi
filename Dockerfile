FROM golang:1.25.7-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
RUN GOBIN=/out go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.1
COPY . .
RUN CGO_ENABLED=0 go build -o /out/miniapi ./cmd/server

FROM alpine:3.22
WORKDIR /app
COPY --from=build /out/miniapi /app/miniapi
COPY --from=build /out/migrate /app/migrate
COPY migrations /app/migrations
ENTRYPOINT ["/app/miniapi"]
