FROM golang:1.23-alpine AS builder
RUN apk add make
WORKDIR /go/src/timescaledb-benchmarker
COPY go.mod .
COPY go.sum .
RUN go mod download
# No copy until this moment in order to take profit of cache in case no deps have changed.
COPY . .
RUN make build

FROM alpine:latest
COPY --from=builder /go/src/timescaledb-benchmarker/bin/out/timescaledb-benchmarker /usr/bin/timescaledb-benchmarker
ENTRYPOINT ["timescaledb-benchmarker"]