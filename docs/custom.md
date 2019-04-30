
# 如何实现自定义镜像

既然是云原生开发方式，最终我们编写的微服务接口需要以容器的方式打包部署。

## 准备`Dockerfile`文件

根据实际情况准备`Dockerfile`，添加开发好的`SQLRestful`文件（`hcl`）到打包镜像中，具体如下所示：

```
# 引入sqlrestful镜像
FROM snz1/sqlrestful

# 把你的HCL配置文件添加到镜像的`/sqlrestful`目录下
ADD <your hcl file> /sqlrestful/

# 把RSA私钥文件添加到镜像的`/sqlrestful`目录下
ADD <rsa privekey file> /sqlrestful/

# 根据生产环境，自定义入口配置参数
ENTRYPOINT ["sqlrestful", "-driver", "postgres", "-dsn", ..., "-jwt-secret", "..."]
```

