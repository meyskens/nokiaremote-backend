FROM alpine:3.8

COPY ./build/bin/nokiaremote /bin/nokiaremote

ENTRYPOINT ["/bin/nokiaremote"]
CMD [ "serve" ]
