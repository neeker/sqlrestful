FROM gitlab.snz1.cn:2008/go/cgobuild:v1.0

ENV TZ=Asia/Shanghai

RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime

ADD . /tmp/sqlrestful

RUN cd /tmp/sqlrestful && \
   CGO_ENABLED=1 GO111MODULE=on \
   go build --tags "linux sqlite_stat4 sqlite_allow_uri_authority sqlite_fts5 sqlite_introspect sqlite_json"

RUN cp -f /tmp/sqlrestful/sqlrestful /usr/local/bin

RUN mkdir -p /test && \
    cp /tmp/sqlrestful/examples/*.hcl /test/

RUN rm -rf /tmp/sqlrestful && \
   mkdir -p /sqlrestful

ENTRYPOINT ["sqlrestful"]

WORKDIR /sqlrestful

