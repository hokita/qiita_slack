# raspberry pi上で動かすのでarm製を使用する。
# FROM golang:1.13
FROM alexellis2/go-armhf:1.7.4

ENV GOBIN /go/bin

RUN go get github.com/nlopes/slack
RUN go get github.com/PuerkitoBio/goquery

WORKDIR /go
ADD . /go

CMD ["go", "run", "main.go"]
