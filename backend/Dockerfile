FROM golang:latest AS build

WORKDIR /compose/backend
RUN go get github.com/mediocregopher/radix
COPY main.go main.go
RUN CGO_ENABLED=0 go build -o backend main.go

FROM scratch
EXPOSE 3333
COPY --from=build /compose/backend/backend /usr/local/bin/backend
CMD ["/usr/local/bin/backend"]
