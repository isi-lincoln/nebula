FROM ubuntu:24.04

# apt grab all the packages we will need
RUN echo 'APT::Install-Suggests "0";' >> /etc/apt/apt.conf.d/00-docker
RUN echo 'APT::Install-Recommends "0";' >> /etc/apt/apt.conf.d/00-docker
RUN DEBIAN_FRONTEND=noninteractive \
  apt-get update \
  && apt-get install -qy \
	wget \
	git \
	make \ 
	protobuf-compiler \
	iproute2 \
	iputils-ping \
	traceroute \
	vim-nox

# set the go environment variables
ENV GOPATH=/root/go
ENV GOBIN=/root/go/bin
ENV PATH="$PATH:/usr/local/go/bin:/root/go/bin"

# download golang 1.22.4
RUN wget https://go.dev/dl/go1.22.4.linux-amd64.tar.gz && tar -C /usr/local -xzf go1.22.4.linux-amd64.tar.gz && rm -rf go1.22.4.linux-amd64.tar.gz*

# get our version of nebula & avoid services
RUN git clone https://github.com/isi-lincoln/nebula.git
WORKDIR nebula
RUN git fetch -a && git checkout avoid-broken

# protobuf / protoc needed binaries for building
RUN go get google.golang.org/protobuf/cmd/protoc-gen-go && go install google.golang.org/protobuf/cmd/protoc-gen-go && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

RUN make

# copy our avoid service binaries to /bin
RUN cp ./avoid/services/actioneer/avoid-actioneer /bin/avoid-actioneer && \
	cp ./avoid/services/client/avoid-client /bin/avoid-client && \
	cp ./avoid/services/manager/avoid-manager /bin/avoid-manager && \
	cp ./avoid/services/relay/avoid-relay /bin/avoid-relay

# create our avoid directory to store configuration file
RUN mkdir -p /etc/avoid

COPY avoid.service.conf /etc/avoid/avoid.conf
