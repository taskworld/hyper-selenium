FROM golang:1.10.3 as builder
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
WORKDIR /go/src/github.com/taskworld/hyper-selenium/
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -vendor-only
COPY agent ./agent
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o hyper-selenium-agent ./agent

FROM selenium/standalone-chrome-debug:3.12.0-cobalt
WORKDIR /hyper-selenium/
COPY --from=builder /go/src/github.com/taskworld/hyper-selenium/hyper-selenium-agent .
CMD ["./hyper-selenium-agent"]
