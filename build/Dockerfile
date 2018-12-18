FROM alpine:3.8

RUN apk upgrade --update --no-cache && apk add curl

USER nobody

ADD build/_output/bin/polaris-operator /usr/local/bin/polaris-operator
