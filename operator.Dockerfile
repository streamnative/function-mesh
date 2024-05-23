FROM alpine:3.20

RUN apk add tzdata --no-cache
RUN apk upgrade --no-cache
ADD bin/function-mesh-controller-manager /manager
