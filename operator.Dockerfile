FROM alpine:3.10

RUN apk add tzdata --no-cache
ADD bin/function-mesh-controller-manager /usr/local/bin/function-mesh-controller-manager
