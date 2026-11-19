# syntax=docker/dockerfile:1

# ---- Build stage ----
FROM golang:1.27-alpine AS builder
WORKDIR /src

# Download dependency dulu (layer ini di-cache selama go.mod/go.sum tidak berubah).
COPY go.mod go.sum ./
RUN go mod download

COPY . .
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api \
 && go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

# ---- Runtime stage ----
# distroless: tanpa shell & package manager -> image kecil dan permukaan serangan minim.
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /out/api /out/migrate /app/

USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/api"]
