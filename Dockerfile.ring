FROM golang:buster AS builder

# Install libpcap
RUN apt-get update && \
  apt-get -y install libpcap0.8 libpcap0.8-dev

# Install pfring
RUN apt-get update && \
  apt-get -y -q install wget lsb-release && \
  wget -q http://apt.ntop.org/16.04/all/apt-ntop.deb && dpkg -i apt-ntop.deb && \
  apt-get clean all && \
  apt-get update && \
  apt-get -y install pfring

# Install pfring zero copy if desired. Not working right now
# RUN apt-get -y install pfring-drivers-zc-dkms

# Set the working directory to ...
WORKDIR /go/src/github.com/traffic-refinery/traffic-refinery/

# Copy the source code and config files
ADD cmd ./cmd/
ADD internal ./internal/
ADD Makefile ./
ADD go.* ./

# Get dependencies
RUN go mod tidy

# Create counters if needed
RUN go run scripts/create_counters.go

# Build TR
RUN make

FROM debian:buster
# Install libpcap
RUN apt-get update && \
  apt-get -y install libpcap0.8 libpcap0.8-dev

# Install pfring
RUN apt-get update && \
  apt-get -y -q install wget lsb-release && \
  wget -q http://apt.ntop.org/16.04/all/apt-ntop.deb && dpkg -i apt-ntop.deb && \
  apt-get clean all && \
  apt-get update && \
  apt-get -y install pfring

WORKDIR /root/
COPY --from=builder /go/src/github.com/traffic-refinery/traffic-refinery/tr /usr/bin/

# Copy configuration files
ADD ./configs config/

# Add folder to drop output.
VOLUME /out

ENTRYPOINT ["/usr/bin/tr"]
CMD ["-name", "trconfig_default.json", "-folder", "/root/config/"]
