FROM golang:1.22.0-alpine3.19 AS builder

# Build the binary statically.
ENV CGO_ENABLED=0

WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN go build -o process .

FROM alpine:3.10.2

# Create kai user.
ENV USER=kai
ENV UID=10001

RUN apk add -U --no-cache ca-certificates
RUN mkdir -p /var/log/app

RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

WORKDIR /app
COPY config.yaml config.yaml

COPY --from=builder /app/process process

RUN chown -R kai:0 /app \
    && chmod -R g+w /app \
    && mkdir /var/log/app -p \
    && chown -R kai:0 /var/log/app \
    && chmod -R g+w /var/log/app

USER kai

CMD ["sh","-c","/app/process 2>&1 | tee -a /var/log/app/app.log"]
