FROM golang:1.26-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build -o /bad-download-cleaner .

FROM alpine:3.24
RUN apk add --no-cache ca-certificates
COPY --from=build /bad-download-cleaner /usr/local/bin/bad-download-cleaner
ENTRYPOINT ["bad-download-cleaner"]
