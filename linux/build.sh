#/bin/bash

BUILD_ORACLE=$1

if [ "$BUILD_ORACLE" == "disabled" ];then
  echo "oracle oci $BUILD_ORACLE"
else
  echo "oracle oci enabled"
fi

if [ "$BUILD_ORACLE" == "disabled" ];then
  CGO_ENABLED=1 GO111MODULE=on \
  go build --tags "linux sqlite sqlite_stat4 sqlite_allow_uri_authority sqlite_fts5 sqlite_introspect sqlite_json"
else
  CGO_ENABLED=1 GO111MODULE=on \
  go build --tags "linux sqlite sqlite_stat4 sqlite_allow_uri_authority sqlite_fts5 sqlite_introspect sqlite_json oracle"
fi

exit $?