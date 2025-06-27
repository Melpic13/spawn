FROM golang:1.22 AS builder
WORKDIR /src
COPY . .
RUN go build -o /out/spawnd ./cmd/spawnd

FROM gcr.io/distroless/base-debian12
COPY --from=builder /out/spawnd /usr/local/bin/spawnd
ENTRYPOINT ["/usr/local/bin/spawnd"]
