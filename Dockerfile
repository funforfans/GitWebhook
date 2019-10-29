FROM golang:1.12.4-alpine3.9 as builder

WORKDIR /go/src/GitWebhook

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o GitWebhook .

FROM znly/protoc as prod

WORKDIR /root/GitWebhook

COPY --from=0 /go/src/GitWebhook .

CMD ["./GitWebhook"]