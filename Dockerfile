FROM golang:1.17-alpine

WORKDIR $GOPATH/src/github.com/xacnio/go-login-auth

COPY . .

RUN go get -d -v ./...

RUN go build

EXPOSE 3000

CMD ["./go-login-auth"]