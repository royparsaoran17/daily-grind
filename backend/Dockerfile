# ---- build ----
FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api
RUN CGO_ENABLED=0 go build -o /out/import-bible ./cmd/import-bible

# ---- run ----
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/api /app/api
COPY --from=build /out/import-bible /app/import-bible
EXPOSE 8080
# Use CMD (not ENTRYPOINT) so you can override with the importer:
#   docker compose run --rm backend /app/import-bible
CMD ["/app/api"]
