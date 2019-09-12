FROM alpine:latest

RUN apk add --update bash curl && rm -rf /var/cache/apk/*

ADD vpn /bin/vpn

EXPOSE 80

CMD ["/bin/vpn"]
