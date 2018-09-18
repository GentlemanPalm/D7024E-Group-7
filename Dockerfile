FROM larjim/kademlialab:latest

RUN mkdir /home/go/src/app
COPY . /home/go/src/app
WORKDIR /home/go/src/app
# RUN apt-get update && apt-get upgrade -y && apt-get -y install iputils-ping
#RUN apt-get update && apt-get upgrade -y










