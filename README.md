# SQL转Restful接口服务

## 设计初衷

一直在使用`Java+SpringBoot`作为微服务生产力工具，通常来讲实现一个Restful微服务接口需要做以下相关工作：

  - ORM映射实现：通常使用MyBatis，需要写Pojo类、Mapper类和SQLProvider类。
  - 服务层实现：需要一个对象管理接口类与一个对象管理实现类。
  - 控制层实现：一个RestConfoller类并注解Rest方法再调用服务层实现。
  - Devops配置：Dockerfile、k8s部署描述文件等等。

这样一个`Restful`接口实现下来至少需要`5`个以上的类，大部分工作是在做转换、校验等语言相关的工作。

我们可以想想从SQL到Restful经历了多少层的实现，而大部分工作是毫无意义的规则代码。。。

> 您可能会建议我使用spring-cloud云原生开发框架，但一样也少不上述过程。

因此一直以来，我一直想要有一个工具可以直接把SQL转成Restful微服务接口，同时它必须是云原生的开发方式：

  - 1、配置化实现：通过简单的配置边可以很方便的实现SQL转Restful接口；
  - 2、执行效率高：不能因为配置和转换减低运行效率；
  - 3、可容器化部署：能方便打包成Docker镜像并运行；
  - 4、多数据库支持：包括oracle、db2、mysql、postgres、hadoop等。

通过此工具可以快速对外提供Restful规范的数据微服务j接口，满足碎片化的数据服务需求应用场景的快速响应。

> 说干就干，于是找到了[sqler](https://github.com/alash3al/sqler)，但是[sqler](https://github.com/alash3al/sqler)仅支持REST而不支持Restful。<br>
>因此我在其基础之上实现了一个完整的SQL转Restful接口的服务工具，在兼容[sqler](https://github.com/alash3al/sqler)配置语法
>的同时进行了Restful配置扩展实现。

感谢开源！


## 开发说明


### JavaScript脚本

在SQLRestful的主要实现由SQL与JavaScript完成，其中JavaScript负责提供与其他微服务接口的交互、SQL返回结果的转换能力。

JavaScript主要用于参数转换（`bind`宏），身份验证实现（`authorizer`宏），应答转换（`transformer`宏）实现，它支持
完整的 ECMAScript 5.1 规范（由 [goja](https://github.com/dop251/goja) 提供实现支持）。

参数转换（`bind`宏），身份验证实现（`authorizer`宏）的JS脚本可以通过变量`$input`可以获取到请求输入参数：

* `$result`表示请求参数JSON对象
* 请求头中的参数通过`http_`开头+头名称（全部小写，`-`被替换成`_`），如有一个请求头叫`x-test-mm`，则通过以下表达式拿到值：

```
$input.http_x_test_mm
```

应答转换（`transformer`宏）的脚本通过变量`$result`可以获取到`exec`宏返回的原始应答JSON对象：

```
transformer = <<JS
(function(){
$new_result = $result
$new_result.trans_test = "13456"
return $new_result  
})()
JS
```

SQLRestful为JS脚本内置了两个默认的HTTP请求函数和一个控制台日志输出函数：

  - fetch
  - call_api
  - log

#### fetch 函数说明

**函数原型**

```
function fetch(URL, {
  method: "HTTP METHOD", //请求方法，如GET、POST、PUT
  headers: { //请求头
    ...
  },
  body: ... //请求体，可以是JSON或字符串。
})
```

**返回结果**

```
{
  "status":     "应答状态文本",
  "statusCode": "HTTP应答码",
  "headers":    "HTTP应答头",
  "body":       "应答内容字符串",
}
```


#### call_api 函数说明

此函数提供后台调用在[应用网关](https://snz1.cn/k8s/javadoc/appgateway/2.%E7%94%A8%E6%88%B7%E6%89%8B%E5%86%8C/ExpSvc.html)中注册的微服务[后台接口](https://snz1.cn/k8s/javadoc/appgateway/2.用户手册/ExpSvc.html#认证模式说明)。

它通过SQLRestful服务配置的 JWT RSA 私钥与 JWT 安全令牌产生 JWT 请求令牌并发起接口请求。

> 具体JWT令牌请求方式参见《[通过网关调用后台服务接口
](https://snz1.cn/k8s/javadoc/appgateway/2.%E7%94%A8%E6%88%B7%E6%89%8B%E5%86%8C/CallApi.html)》中的说明。

**函数原型**

```
function call_api(URL, {
  method: "HTTP METHOD", //请求方法，如GET、POST、PUT
  headers: { //请求头
    ...
  },
  body: ... //请求体，可以是JSON或参数内容。
})
```

**返回结果**

正常情况下 call_api 函数直接返回接口的JSON对象，只有在请求出错的情况下返回如下定义：


```
{
  "status":     "应答状态文本",
  "statusCode": "HTTP应答码",
  "headers":    "HTTP应答头",
  "body":       "应答内容字符串",
}
```

如果请求的接口应答内容不能转换为JSON对象则返回与`fetch`函数相同的应答：

### 内置的接口说明

#### 心跳检测

* 接口地址：/health
* 请求方法：GET
* 应答格式：JSON

```json
```

#### 清理所有缓存

* 接口地址：/clear_caches
* 请求方法：POST
* 应答格式：JSON

```json
```

#### swagger2.0文档接口

* 接口地址：/v2/api-docs
* 请求方法：GET
* 应答格式：JSON

```json
```

## 运行SQLRestful服务

准备好HCL配置文件以后即可对外提供微服务接口了，你可以独立运行docker镜像，也可以使用DevOPS流程部署到容器环境中。

### 查看帮助

```
docker run -ti --rm snz1/sqlrestful --help
```

参数说明：

```
Usage of sqlrestful:
  -config string
        缺省的配置文件路径（多个文件使用逗号分隔） (default "./*.hcl")
  -driver string
        SQL类型 (default "postgres")
  -dsn string
        SQL数据源配置 (default "user=postgres password= dbname=postgres sslmode=disable connect_timeout=3")
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
  -port string
        HTTP监听端口 (default ":80")
  -redis string
        Redis连接：redis://:password@<redis host>:6379/0
  -sep string
        SQL分隔符 (default "---\\\\--")
  -workers int
        工作线程数量 (default 1)
```

### 数据库驱动及连接串

| 数据库 | 连接串 |
---------| ------ |
| `mysql`| `usrname:password@tcp(server:port)/dbname?option1=value1&...`|
| `postgres`| `postgresql://username:password@server:port/dbname?option1=value1`|
|           | `user=<dbuser> password=<password> dbname=<dbname> sslmode=disable connect_timeout=3 host=<db host>` |
| `sqlite3`| `/path/to/db.sqlite?option1=value1`|
| `sqlserver` | `sqlserver://username:password@host/instance?param1=value&param2=value` |
|             | `sqlserver://username:password@host:port?param1=value&param2=value`|
|             | `sqlserver://sa@localhost/SQLExpress?database=master&connection+timeout=30`|
| `mssql` | `server=localhost\\SQLExpress;user id=sa;database=master;app name=MyAppName`|
|         | `server=localhost;user id=sa;database=master;app name=MyAppName`|
|         | `odbc:server=localhost\\SQLExpress;user id=sa;database=master;app name=MyAppName` |
|         | `odbc:server=localhost;user id=sa;database=master;app name=MyAppName` |
| `hdb` (SAP HANA) |   `hdb://user:password@host:port` |
| `clickhouse` (Yandex ClickHouse) |   `tcp://host1:9000?username=user&password=qwerty&database=clicks&read_timeout=10&write_timeout=20&alt_hosts=host2:9000,host3:9000` |

### 配置JWT请求令牌参数

需要把应用的 RSA 私钥文件放到镜像的文件系统中，然后在命令行中加入`jwt-keyfile`、`jwt-secret`、`jwt-expires`参数：

```
-jwt-keyfile "/sqlrestful/app.pem" -jwt-secret "***********" -jwt-expires=3600
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

## 计划功能

 - [x] 实现Redis缓存配置，为restful接口实现缓存接口。
 - [x] 实现标准的swagger-ui文档接口(/v2/api-docs)。
 - [ ] 加入oracle、db2等商用数据库支持。
 - [x] 完善在JS中发起JWT请求令牌请求其他接口。
 - [x] 编写JS脚本及内置函数说明文档。

