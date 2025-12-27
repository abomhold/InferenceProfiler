# Build
FROM golang:1.25 AS builder
WORKDIR /app
COPY ./src ./src
COPY go.mod .
COPY Makefile .
RUN CGO_ENABLED=1 GOOS=linux make build

# Setup
FROM ubuntu:24.04
ENV DEBIAN_FRONTEND=noninteractive
RUN mkdir -p /profiler-output
COPY --from=builder /app/bin/profiler /usr/local/bin/profiler
ENTRYPOINT ["/usr/local/bin/profiler", "-o", "/profiler-output", "-t", "5000"]

# User
WORKDIR /app
RUN apt-get update && apt-get install -y --no-install-recommends sysbench \
    && rm -rf /var/lib/apt/lists/*
CMD ["sysbench", "--test=cpu", "--cpu-max-prime=2000", "--max-requests=400", "run"]