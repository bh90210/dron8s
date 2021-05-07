FROM alpine:3.13.5
ADD dron8s /bin/
RUN apk -Uuv add ca-certificates
ENTRYPOINT /bin/dron8s