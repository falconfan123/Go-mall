FROM golang:1.25.8

WORKDIR /workspace
COPY common ./common
COPY dal ./dal
COPY services ./services
COPY test/rpc ./test/rpc

WORKDIR /workspace/test/rpc
ENV GOWORK=off
CMD ["go", "test", "-json", "-count=1", "./..."]
