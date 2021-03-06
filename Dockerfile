FROM golang:1.13.5-alpine3.10
WORKDIR /build
COPY . .
RUN apk --no-cache add build-base
RUN GOOS=linux go build -a -ldflags '-s -w -extldflags "-static"' -o launchpad .

FROM alpine:3.10
RUN apk --no-cache add ca-certificates
WORKDIR /launchpad
COPY --from=0 /build/launchpad .
CMD ["./launchpad", "runner"]