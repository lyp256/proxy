FROM golang:1 as build
COPY . /src
RUN cd /src && CGO_ENABLED=0 make build
RUN chmod +x /src/build/proxy

FROM alpine:latest as prod
COPY --from=build /src/build/proxy /usr/bin/
CMD proxy
