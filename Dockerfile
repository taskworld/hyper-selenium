FROM golang:1.10.3 as builder
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
WORKDIR /go/src/github.com/taskworld/hyper-selenium/
COPY Gopkg.toml Gopkg.lock ./
ENV CGO_ENABLED=0
ENV GOOS=linux
RUN dep ensure -vendor-only
COPY cmd ./cmd
COPY pkg ./pkg
RUN mkdir -p build && go build -a -v -installsuffix cgo -o ./build/hyper-selenium-agent ./cmd/hyper-selenium-agent

FROM selenium/standalone-chrome-debug:3.12.0-cobalt AS env
RUN sudo apt-get update && sudo apt-get install -y ffmpeg gpac && sudo rm -rf /var/lib/apt/lists/*
RUN sudo mkdir -p /videos/
WORKDIR /hyper-selenium/
ENV SCREEN_WIDTH=1280
ENV SCREEN_HEIGHT=1024

FROM env
COPY --from=builder /go/src/github.com/taskworld/hyper-selenium/build/hyper-selenium-agent .
CMD ["./hyper-selenium-agent"]
