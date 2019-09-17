FROM alpine:3.7

RUN apk update && \
    apk upgrade && \
    apk --no-cache add curl jq file

VOLUME [ /ssd ]
WORKDIR /ssd
EXPOSE 26656 26657
ENTRYPOINT ["/usr/bin/wrapper.sh"]
CMD ["start"]
STOPSIGNAL SIGTERM

COPY wrapper.sh /usr/bin/wrapper.sh
RUN chmod +x /usr/bin/wrapper.sh

