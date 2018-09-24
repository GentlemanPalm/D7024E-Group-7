FROM larjim/kademlialab:latest

RUN mkdir /home/go/src/app
COPY . /home/go/src/app
COPY d7024e /home/go/src/d7024e
WORKDIR /home/go/src/app
RUN mkdir /home/go/src/NetworkMessage
ENV GOPATH /home/go
ENV PATH="${GOPATH}/bin:${PATH}"
# RUN apt-get update && apt-get upgrade -y && apt-get -y install iputils-ping
#RUN apt-get update && apt-get upgrade -y
RUN protoc --go_out=${GOPATH}/src/NetworkMessage *.proto
RUN CGO_ENABLED=0 GOOS=linux GOARCH=386 /usr/local/go/bin/go build -o main .
#RUN echo $GOPATH
#CMD ["/usr/local/go/bin/go","run","main.go"]
CMD ["./run.sh"]
