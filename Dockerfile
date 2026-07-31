# Multi-stage build: compile the frontend, embed it into a static Go binary,
# then ship a minimal runtime image.

# ---- frontend ----
FROM node:20-alpine AS web
WORKDIR /src
COPY web/package.json web/package-lock.json ./web/
RUN cd web && npm ci
COPY web ./web
COPY internal/web ./internal/web
# vite build.outDir is ../internal/web/dist, so this writes to /src/internal/web/dist
RUN cd web && npm run build

# ---- binary ----
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web /src/internal/web/dist ./internal/web/dist
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -tags embed \
    -ldflags "-s -w -X github.com/3sarojbhattarai/gantry/internal/cli.version=${VERSION}" \
    -o /gantry ./cmd/gantry

# ---- runtime ----
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /gantry /usr/local/bin/gantry
EXPOSE 8080
# Binds 0.0.0.0 so the port is reachable outside the container. SECURITY: this
# grants full, unauthenticated Docker control to anyone who can reach the
# published port and requires mounting the host's Docker socket — see README.
ENTRYPOINT ["gantry"]
CMD ["serve", "--addr", "0.0.0.0:8080"]
