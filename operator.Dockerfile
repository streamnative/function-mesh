FROM alpine:3.14

RUN apk add tzdata --no-cache
ADD bin/function-mesh-controller-manager /manager
