FROM golang:1.26-alpine AS build
WORKDIR /src

COPY . .

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