FROM qnib/uplain-golang

WORKDIR /usr/local/src/github.com/qnib/havener
COPY ./main.go /usr/local/src/github.com/qnib/havener/
COPY ./v1 /usr/local/src/github.com/qnib/havener/v1
COPY vendor/vendor.json vendor/vendor.json
RUN govendor fetch -v +m \
 && govendor install

FROM qnib/uplain-init

COPY --from=0 /usr/local/bin/havener /usr/local/bin/
COPY ./v1/static /usr/share/havener/static
CMD ["havener"]
