FROM golang:alpine

RUN \
	apk add --update \
		python \
		python-dev \
		py-pip \
		build-base \
    linux-headers \
	&& \
	pip install dumb-init \
	&& \
	rm -rf /var/cache/apk/* && \
	:

RUN apk update && \
    apk add curl make git && \
    apk del curl && \
    rm -rf /var/cache/apk/*

ENTRYPOINT ["dumb-init"]
CMD ["/bin/sh"]
