FROM alpine:3.19

RUN apk add tzdata --no-cache
ADD bin/function-mesh-controller-manager /manager
