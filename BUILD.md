# 编译说明

## `Linux`镜像编译

### 完整功能版编译

```sh
docker build -t gitlab.snz1.cn:2008/go/sqlrestful:v0.7ex .
```

### 无`oci`版编译

```sh
docker build --build-arg ORACLE=disabled -t gitlab.snz1.cn:2008/go/sqlrestful:v0.7ex .
```

## `Windows`版编译

```
win\build.bat
```
