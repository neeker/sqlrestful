# 编译说明

## `Linux`镜像编译

### 无`oci8`编译

```sh
docker build -t snz1/sqlrestful .
```

### 加上`oci8`编译

```sh
docker build --build-arg GOBUILD_TAGS=oracle -t snz1/sqlrestful .
```

## `Windows`版编译

```cmd
win\build.bat
```
