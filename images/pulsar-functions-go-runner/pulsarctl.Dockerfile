ARG BASE_IMAGE=pulsar-functions-pulsarctl-runner-base:latest

FROM ${BASE_IMAGE}

WORKDIR /pulsar

USER $USER
