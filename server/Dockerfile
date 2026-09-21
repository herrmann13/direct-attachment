# ---- Build stage ---------------------------------------------------------
FROM golang:1.22-alpine AS build

WORKDIR /src

# Cache dependencies before copying sources.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# ---- Runtime stage -------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/server /server

EXPOSE 8080
ENTRYPOINT ["/server"]
