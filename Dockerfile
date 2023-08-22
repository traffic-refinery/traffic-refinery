FROM golang:bookworm AS builder

# Install libpcap
RUN apt-get update && \
  apt-get -y install libpcap0.8 libpcap0.8-dev

# Set the working directory to ...
WORKDIR /go/src/github.com/traffic-refinery/traffic-refinery/

# Copy the source code and config files
ADD cmd ./cmd/
ADD internal ./internal/
ADD scripts ./scripts/
ADD Makefile ./
ADD go.* ./

# Get dependencies
RUN go mod tidy

# Create counters if needed
RUN go run scripts/create_counters.go

# Build TR
RUN make

FROM debian:bookworm
# Install libpcap
RUN apt-get update && \
  apt-get -y install libpcap0.8 libpcap0.8-dev

WORKDIR /root/
COPY --from=builder /go/src/github.com/traffic-refinery/traffic-refinery/tr /usr/bin/

# Copy configuration files
ADD ./configs config/

# Add folder to drop output.
VOLUME /out

ENTRYPOINT ["/usr/bin/tr"]
CMD ["-name", "trconfig_default.json", "-folder", "/root/config/"]
