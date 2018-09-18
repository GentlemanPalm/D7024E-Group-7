FROM larjim/kademlialab:latest

RUN mkdir /home/go/src/app
COPY . /home/go/src/app
COPY d7024e /home/go/src/d7024e
WORKDIR /home/go/src/app
ENV GOPATH /home/go/
# RUN apt-get update && apt-get upgrade -y && apt-get -y install iputils-ping
#RUN apt-get update && apt-get upgrade -y
RUN CGO_ENABLED=0 GOARCH=386 /usr/local/go/bin/go build -o main .
#RUN echo $GOPATH
CMD ["/usr/local/go/bin/go","run","main.go"]
