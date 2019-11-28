
## 查看帮助

```
docker run -ti --rm snz1/sqlrestful --help
```

参数说明：

```
Usage of sqlrestful:
  -author string
        维护人员 (default "痞子飞猪")
  -base string
        服务地址 (default "/")
  -config string
        缺省的配置文件路径（多个文件使用逗号分隔） (default "./*.hcl")
  -desc string
        功能描述 (default "SQLRestful，您的云原生应用生产力工具！")
  -driver string
        SQL类型
  -dsn string
        SQL数据源配置
  -email string
        联系邮箱 (default "13317312768@qq.com")
  -hdb.protocol.trace
        enabling hdb protocol trace
  -hdb.sqlTrace
        enabling hdb sql trace
  -jwt-expires int
        JWT安全令牌 (default 1800)
  -jwt-keyfile string
        RSA私钥文件（PEM格式） (default "./app.pem")
  -jwt-secret string
        JWT安全令牌
  -name string
        服务名称 (default "SQLRestful")
  -port string
        HTTP监听端口 (default ":80")
  -redis string
        Redis连接：redis://:password@<redis host>:6379/0
  -sep string
        SQL分隔符 (default "---\\\\--")
  -ver string
        实现版本 (default "1.0")
  -workers int
        工作线程数量 (default 1)
```


### 运行服务

**运行指定目录下的配置**

```
docker run -ti --rm \
  -v /path/of/your/sqlrestful:/sqlrestful \
  -v /pathof/your/app.pem:/sqlrestful/app.pem:ro \
  -p 80:80 \
  snz1/sqlrestful \
  -driver "postgres" \
  -dsn "postgesql://username:password@server:port/dbname?sslmode=disable&connect_timeout=3" \
  -redis "redis://:password@server:port/0" \
  -jwt-keyfile "./app.pem" \
  -jwt-secret "**********" \
  -jwt-expires 3600
```

**运行示例目录的配置**

```
docker run -ti --rm \
  -p 80:80 \
  snz1/sqlrestful \
  -driver "postgres" \
  -dsn "postgesql://username:password@server:port/dbname?sslmode=disable&connect_timeout=3"
  -config "/test/*.hcl"
```
