# ---- Build stage ---------------------------------------------------------
FROM golang:1.22-alpine AS build

WORKDIR /src

# Cache dependencies before copying sources. The build context is the
# repository root, so paths are prefixed with server/.
COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# ---- Runtime stage -------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/server /server

EXPOSE 8080
ENTRYPOINT ["/server"]
