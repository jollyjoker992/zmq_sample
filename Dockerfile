FROM golang:1.12

WORKDIR $GOPATH/src/github.com/jollyjoker992/zmq_sample

ARG username
ARG password

# libzmq
RUN apt-get update

RUN DEBIAN_FRONTEND=noninteractive apt-get install -y git build-essential libtool autoconf automake pkg-config unzip libkrb5-dev

RUN cd /tmp && git clone git://github.com/jedisct1/libsodium.git && cd libsodium && git checkout e2a30a && ./autogen.sh && ./configure && make check && make install && ldconfig

RUN cd /tmp && git clone --depth 1 git://github.com/zeromq/libzmq.git && cd libzmq && ./autogen.sh && ./configure && make

RUN cd /tmp/libzmq && make install && ldconfig

RUN rm /tmp/* -rf

## main program
RUN git clone https://${username}:${password}@github.com/jollyjoker992/zmq_sample.git

COPY . .

RUN go get -d -v ./...

RUN cd server && go install

EXPOSE 5555

CMD ["server"]
