FROM golang:1.17-alpine AS build_deps

RUN apk add --no-cache git

WORKDIR /workspace

COPY go.mod .
COPY go.sum .

RUN go mod download

FROM build_deps AS build

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dron8s -ldflags '-w -extldflags "-static"' .

FROM gcr.io/distroless/static

COPY --from=build /workspace/dron8s /bin/

ENTRYPOINT ["dron8s"]
