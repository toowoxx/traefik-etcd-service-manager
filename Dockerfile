from golang:1.18-buster as builder
add . /src
workdir /src
run CGO_ENABLED=0 go build -o app

from alpine
copy --from=builder /src/app /app/app
entrypoint ["/app/app"]
