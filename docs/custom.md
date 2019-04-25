
### 自定义镜像

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