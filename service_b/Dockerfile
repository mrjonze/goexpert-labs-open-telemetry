FROM golang:1.22 as service-b
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o goapp

FROM alpine:latest
WORKDIR /app
COPY --from=service-b /app/goapp .
EXPOSE 8081
CMD ["./goapp"]