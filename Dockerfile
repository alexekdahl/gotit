FROM golang:latest as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -mod=readonly -ldflags="-s -w" -gcflags=all=-l -trimpath=true -o ./bin/gotit ./cmd/gotit/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/bin/gotit ./bin/

COPY --from=builder /app/authorized_keys /root/authorized_keys

EXPOSE 8080 2222

CMD ["./bin/gotit"]
