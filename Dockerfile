FROM golang:1.14-alpine

# Set timezone
ENV TZ "Asia/Tokyo"

# Install tools required to build the project
RUN apk add --no-cache ca-certificates \
  curl \
  ffmpeg \
  git \
  make \
  rtmpdump \
  tzdata

ARG PROJECT_PATH=/go/src/github.com/yyoshiki41/radigo
WORKDIR ${PROJECT_PATH}
COPY . ${PROJECT_PATH}/

# Install deps
RUN make installdeps
# Build the project binary
RUN make build-4-docker

# Set default output dir
WORKDIR /
VOLUME ["/output"]

ENTRYPOINT ["/bin/radigo"]
CMD ["--help"]
