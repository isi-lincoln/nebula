# Overview

Building docker containers for milcom-demo.

The container doesnt have a command step, so when deploying the container, the command must be present.

## Building

```
sudo docker build -t docker.io/isi-lincoln/avoid-demo:v0.0.0 -f Dockerfile .
```

## Running

```
sudo docker run -v /etc/etcd:/etc/etcd -it isilincoln/avoid-demo:v0.0.0 /bin/bash
```
