FROM golang:1.23-alpine AS builder

ARG CGO_ENABLED=0
ARG GOOS=linux
ARG GOARCH=amd64
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

WORKDIR /app/cmd/
RUN go get -v ./... \
  && go install -v ./... \
  && go build -v -o executor

FROM scratch AS production

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/cmd/executor /app/executor

ENTRYPOINT [ "/app/executor" ]
