FROM golang:1.16 as build

COPY ./ /go/src/github.com/meyskens/nokiaremote-backend
WORKDIR /go/src/github.com/meyskens/nokiaremote-backend

RUN PKG_OS=linux make packages

FROM ubuntu:20.04

COPY --from=build /go/src/github.com/meyskens/nokiaremote-backend/build/bin/nokiaremote /bin/nokiaremote
COPY --from=build /go/src/github.com/meyskens/nokiaremote-backend/static /opt/nokiaremote

WORKDIR /opt/nokiaremote

ENTRYPOINT ["/bin/nokiaremote"]
CMD [ "serve" ]
