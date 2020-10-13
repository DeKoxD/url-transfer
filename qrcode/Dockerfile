FROM golang:latest AS build

WORKDIR /compose/qrcode
COPY main.go main.go
RUN go get github.com/skip2/go-qrcode
RUN CGO_ENABLED=0 go build -o qrcode main.go

FROM scratch
EXPOSE 3333
COPY --from=build /compose/qrcode/qrcode /usr/local/bin/qrcode
CMD ["/usr/local/bin/qrcode"]
