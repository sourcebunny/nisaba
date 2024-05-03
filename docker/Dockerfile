# Start from a base image with Go installed
FROM golang:alpine as builder

# Install git and other dependencies
RUN apk update && apk add --no-cache git curl build-base

# Set the working directory inside the container
WORKDIR /app

# Clone and build Nisaba
RUN git clone https://github.com/sourcebunny/nisaba.git /app/nisaba
WORKDIR /app/nisaba
RUN go get github.com/thoj/go-ircevent && \
    go build -o nisaba .

# Download llamafile
RUN curl -L -o /app/llamafile https://github.com/Mozilla-Ocho/llamafile/releases/download/0.8.1/llamafile-0.8.1 && \
    chmod +x /app/llamafile

# Start from a fresh base image to reduce size
FROM alpine:latest
COPY --from=builder /app/nisaba/nisaba /app/nisaba
COPY --from=builder /app/llamafile /app/llamafile
WORKDIR /app

# Copy the entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Expose the port used by llamafile
EXPOSE 8080

ENTRYPOINT ["/entrypoint.sh"]
