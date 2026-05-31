FROM golang:1.23-alpine AS builder

WORKDIR /build

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /collector ./cmd/collector

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /collector /collector

USER nonroot:nonroot

ENTRYPOINT ["/collector"]
