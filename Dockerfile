FROM golang:1.9.1
WORKDIR /go/src/github.com/weberc2/maze
COPY . .
RUN go get -d -v .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM multiarch/alpine:armhf-edge
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/github.com/weberc2/maze/ .
CMD ["./app"]
