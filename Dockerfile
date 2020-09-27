FROM alpine:latest AS downloader
ARG kubectl_url="https://dl.k8s.io/v1.18.9/kubernetes-client-linux-amd64.tar.gz"
ARG kubectl_sha512="e3a5cb14ac277959254dd64bfa0f5d6f09ce338d3bef9865bd5fa1cf828d56468de4d92a03b538042b6d13703403e1f54a54df574f2a12e9800da19939445eb0"
RUN apk add curl
USER 1000
WORKDIR /tmp
RUN curl -L -o kubectl.tgz ${kubectl_url}
RUN echo "${kubectl_sha512}  kubectl.tgz" | tee /tmp/kubectl.sha512sum
RUN sha512sum -c kubectl.sha512sum
RUN tar -xzf kubectl.tgz

FROM golang:1.15-alpine AS builder
RUN apk add --no-cache git
ADD . /src
RUN chown -R 1000:users /src
USER 1000
WORKDIR /src
ENV GOCACHE=/tmp/.go-cache
RUN go build -o cyklistctl ./cmd/cyklistctl

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=downloader --chown=0:0 /tmp/kubernetes/client/bin/kubectl /usr/local/bin/kubectl
COPY --from=builder --chown=0:0 /src/cyklistctl /usr/local/bin/cyklistctl
USER 1000
ENTRYPOINT ["/usr/local/bin/cyklistctl"]
