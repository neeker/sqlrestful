
### 配置JWT请求令牌参数

需要把应用的 RSA 私钥文件放到镜像的文件系统中，然后在命令行中加入`jwt-keyfile`、`jwt-secret`、`jwt-expires`参数：

```
-jwt-keyfile "/sqlrestful/app.pem" -jwt-secret "***********" -jwt-expires=3600
```