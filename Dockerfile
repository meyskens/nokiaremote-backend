FROM ubuntu

COPY ./build/bin/nokiaremote /bin/nokiaremote

ENTRYPOINT ["/bin/nokiaremote"]
CMD [ "serve" ]
