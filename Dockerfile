FROM node:20-alpine AS fe
WORKDIR /fe
COPY frontend/package.json ./
RUN npm install --no-audit --no-fund
COPY frontend/ ./
RUN npm run build
RUN echo "=== fe dist contents ===" && ls -la /fe/dist/ && ls -la /fe/dist/assets/ || true

FROM golang:1.26-alpine AS build
WORKDIR /src
COPY . .
COPY --from=fe /fe/dist ./internal/web/dist
RUN echo "=== go stage dist contents ===" && ls -la /src/internal/web/dist/ && ls -la /src/internal/web/dist/assets/ || true
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -trimpath \
    -o /out/pulse ./cmd/pulse

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/pulse /pulse
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/pulse"]
