FROM debian:8.0
RUN apt-get update
RUN apt-get -q -y install git
ADD playground-server /
WORKDIR /
ENTRYPOINT [ "/playground-server" ]
