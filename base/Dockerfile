FROM ubuntu:14.10

RUN apt-get update
RUN apt-get install -y sudo
RUN apt-get install -y gcc g++ php5-cli ruby python golang-go nodejs perl npm
RUN npm install -g underscore jquery
ENV NODE_PATH /usr/local/lib/node_modules/
ADD run.sh .
ADD entrypoint.sh .

ONBUILD ADD . /code

ENTRYPOINT ["/bin/bash", "/entrypoint.sh"]
