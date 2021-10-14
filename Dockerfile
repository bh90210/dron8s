FROM gcr.io/distroless/static
ADD dron8s /bin/
ENTRYPOINT ["dron8s"]
