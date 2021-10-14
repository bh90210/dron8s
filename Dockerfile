FROM golang:1.17-alpine AS build

RUN apk add --no-cache git

WORKDIR /workspace
COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dron8s -ldflags '-w -extldflags "-static"' .

FROM gcr.io/distroless/static

COPY --from=build /workspace/dron8s /bin/

ENTRYPOINT ["dron8s"]
