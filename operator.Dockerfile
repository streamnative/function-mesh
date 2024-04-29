FROM alpine:3.19

RUN apk add tzdata --no-cache
RUN apk upgrade --no-cache
ADD bin/function-mesh-controller-manager /manager
