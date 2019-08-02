FROM ubuntu

COPY ./build/bin/nokiaremote /bin/nokiaremote
COPY ./static /opt/nokiaremote

WORKDIR /opt/nokiaremote

ENTRYPOINT ["/bin/nokiaremote"]
CMD [ "serve" ]
