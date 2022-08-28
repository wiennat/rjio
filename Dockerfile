FROM golang:1.19 as builder

WORKDIR /src

COPY ./go.mod ./go.sum Makefile ./
RUN make deps

# Import the code from the context.
COPY ./ ./

RUN make build

FROM ubuntu AS final

RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates
# RUN apk --update upgrade && \
#    apk add sqlite ca-certificates && \
#    rm -rf /var/cache/apk/*
# See http://stackoverflow.com/questions/34729748/installed-go-binary-not-found-in-path-on-alpine-linux-docker
# RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

COPY --from=builder /src/dist/rjio* /app/
EXPOSE 3000

ENTRYPOINT ["/app/rjio"]
