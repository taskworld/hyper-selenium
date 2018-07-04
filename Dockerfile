FROM golang:1.10.3 as builder
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
WORKDIR /go/src/github.com/taskworld/hyper-selenium/
COPY Gopkg.toml Gopkg.lock ./
ENV CGO_ENABLED=0
ENV GOOS=linux
RUN dep ensure -vendor-only
COPY agent ./agent
RUN mkdir -p build && go build -a -v -installsuffix cgo -o build/hyper-selenium-agent ./agent

FROM selenium/standalone-chrome-debug:3.12.0-cobalt
RUN sudo apt-get update && sudo apt-get install -y ffmpeg gpac && sudo rm -rf /var/lib/apt/lists/*
WORKDIR /hyper-selenium/
COPY --from=builder /go/src/github.com/taskworld/hyper-selenium/build/hyper-selenium-agent .
ENV SCREEN_WIDTH=1280
ENV SCREEN_HEIGHT=1024
CMD ["./hyper-selenium-agent"]
