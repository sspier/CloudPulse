FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# build the binary named "bootstrap" for AWS Lambda custom runtime
RUN CGO_ENABLED=0 GOOS=linux go build -o bootstrap ./apps/runner

FROM public.ecr.aws/lambda/provided:al2023
COPY --from=builder /app/bootstrap /bootstrap
ENTRYPOINT [ "/bootstrap" ]
