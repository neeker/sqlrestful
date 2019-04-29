#/bin/bash

BUILD_ORACLE=$1

if [ "$BUILD_ORACLE" == "disabled" ];then
  echo "oracle oci $BUILD_ORACLE"
else
  echo "oracle oci enabled"
fi

if [ "$BUILD_ORACLE" == "disabled" ];then
  sed -i "s,github.com/mattn/go-oci8,//github.com/mattn/go-oci8,g" /tmp/sqlrestful/*.mod
fi

if [ "$BUILD_ORACLE" == "disabled" ];then
    sed -i "s,_ \"github.com/mattn/go-oci8\",//_ \"github.com/mattn/go-oci8\",g" /tmp/sqlrestful/prep.go
fi

exit 0
