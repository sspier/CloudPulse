# build stage ----------------------------------------------------
# use alpine-based golang image for a smaller build environment
# base image to build the cloudpulse api binary
FROM golang:1.22-alpine AS builder

# install basic certs so the built binary can make tls calls
# install the certificates the app needs to trust HTTPS connections
# creates /etc/ssl/certs/ca-certificates.crt
RUN apk add --no-cache ca-certificates

# set build workspace
# set working directory inside the container
# this is where the source code will be copied
WORKDIR /app

# copy module files first to take advantage of docker layer caching
# download dependencies
# if only source code changes, dependencies are not re-downloaded
COPY go.mod go.sum ./
RUN go mod download

# copy the rest of the project
COPY apps/api ./apps/api

# build the cloudpulse api binary
# disable cgo to produce a static binary that runs cleanly on alpine
WORKDIR /app/apps/api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/cloudpulse-api

# runtime stage --------------------------------------------------
# minimal alpine base for production
FROM alpine:3.19

# create a non-root user for better security
RUN adduser -D -g '' appuser
USER appuser

# copy the compiled api binary from the builder stage
COPY --from=builder /bin/cloudpulse-api /cloudpulse-api

# default api port
ENV PORT=8080
EXPOSE 8080

# run the api server
ENTRYPOINT ["/cloudpulse-api"]

