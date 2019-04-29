FROM gitlab.snz1.cn:2008/go/cgobuild:v2.0

ARG ORACLE=enabled

ENV TZ=Asia/Shanghai

RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime

ADD examples/sqlite.hcl /examples/

ADD *.go /tmp/sqlrestful/

ADD *.mod /tmp/sqlrestful/

ADD *.sh /tmp/sqlrestful/

RUN chmod +x /tmp/sqlrestful/build.sh

RUN echo "build oracle oci $ORACLE"

RUN /tmp/sqlrestful/build.sh "$ORACLE"

RUN cd /tmp/sqlrestful && \
   CGO_ENABLED=1 GO111MODULE=on \
   go build --tags "linux sqlite_stat4 sqlite_allow_uri_authority sqlite_fts5 sqlite_introspect sqlite_json"

RUN cp -f /tmp/sqlrestful/sqlrestful /usr/local/bin

ADD /swagger2 /swagger2

RUN rm -rf /tmp/sqlrestful

WORKDIR /sqlrestful

ENTRYPOINT [ "sqlrestful" ]

