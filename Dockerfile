FROM alpine:latest AS downloader
ARG kubectl_url="https://dl.k8s.io/v1.15.3/kubernetes-client-linux-amd64.tar.gz"
ARG kubectl_sha512="93049dcadbe401fc2ed2f53ace598aa4bd183142ec2b7451d2a53e61c4bbc64f393639df166853cbf62c361839f87a386015c0be6b1a9a4d3c9fa84564d376ef"
RUN apk add curl
USER 1000
WORKDIR /tmp
RUN curl -L -o kubectl.tgz ${kubectl_url}
RUN echo "${kubectl_sha512}  kubectl.tgz" | tee /tmp/kubectl.sha512sum
RUN sha512sum -c kubectl.sha512sum
RUN tar -xzf kubectl.tgz

FROM golang:1.12-alpine AS builder
RUN apk add --no-cache git make
ADD . /src
RUN chown -R 1000:users /src
USER 1000
WORKDIR /src
ENV GOCACHE=/tmp/.go-cache
RUN make

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=downloader --chown=0:0 /tmp/kubernetes/client/bin/kubectl /usr/local/bin/kubectl
COPY --from=builder --chown=0:0 /src/cyklistctl /usr/local/bin/cyklistctl
USER 1000
ENTRYPOINT ["/usr/local/bin/cyklistctl"]
