# 基本语法结构

`SQLRestful`由**服务配置**、**接口定义**两部分构成，其中服务配置包括名称、概述、版本、作者、数据库连接、缓存连接等，接口定义又分为两层结构：**路径定义**与**方法定义**。

## 服务配置

```hcl
//描述信息
_meta {

  //名称
  name = "SQLRestful"

  //版本
  version = "1.0"

  //描述
  desc = "SQLRestful，您的云原生应用生产力工具！"

  //作者
  author {
    //姓名
    name = "痞子飞猪"
    //邮箱
    email  = "13317312768@qq.com"
  }

  //数据库配置
  db {
    //数据库驱动名
    driver = "sqlite3"
    //数据库连接
    dsn = "sqlte3.db?create=true"
    //Redis缓存连接
    redis = "redis://:password@<redis host>:6379/0"
  }

  //安全配置
  security {
    //统一用户安全服务地址
    api = "https://snz1.cn/test/xeai"
    //用户标识头
    from = "iv-user"
    //用户标识类型
    idtype = "uname"
    //组织域范围
    scope = "employee"
  }

  //应用身份配置：JWT请求令牌
  jwt = {
    //应用私钥(PEM格式)
    rsa = <<PEM
    ....
    PEM

    //应用密钥(安全令牌)
    secret = ""

    //请求令牌超时事件(秒)
    expires = 1800
  }

  //常量定义
  const {
    //...
  }

  //消息插件配置
  mq {
    //消息插件
    driver = "stomp"

    //消息连接
    url = "tcp://stomp_host:61613"
  }

}
```

### 服务配置项说明

#### `name`

服务名称定义

#### `version`

服务实现版本

#### `desc`

服务实现描述

#### `author`

服务维护人员，包括名称（`name`）、邮件（`email`）、网址（`url`）属性。

#### `db`

服务数据库连接配置，包括驱动（`driver`）、连接（`dsn`）、Redis缓存（`redis`）属性。

#### security

统一安全服务接口配置，包括服务接口地址（`api`）、用户标识请求头（`from`）、用户标识类型（`idtype`）、组织域范围（`scope`）属性。

#### jwt

应用身份配置，包括`RSA`私钥（`rsa`）、协商密钥（`secret`）、令牌超时时间（`expires`）等属性。

#### const

全局常量配置表，键值为`js`表达式。

#### mq

消息服务配置，包括消息插件名称（`driver`）、消息连接地址（`url`）等属性。

## 接口定义

### 路径定义语法

```hcl

//接口名称，不能包含“\”、“/”、“:”等特殊字符
api {

  //接口地址，省略时使用接口定义名称作为接口地址，可使用“:变量”方式定义路径变量
  path = "/path/of/object_items"

  //服务接口的分类标签（可忽略）
  tags = ["标签"]

  //摘要描述（可忽略）
  desc = ""

  //常量定义
  const {
    ...
  }

  //接口GET请求方法
  get {
   ... //方法定义
  }

  //POST请求方法
  post {
    ... //方法定义
  }

  //PUT请求方法
  put {
    ... //方法定义
  }

  //PATCH请求方法
  patch {
    ... //方法定义
  }

  //DELETE请求方法
  delete {
    ... //方法定义
  }

}

```

> 当**路径定义**只有一种方法实现可以直接使用`方法定义`语法来简化接口定义：

```hcl

api_name {

  path = "/path/of/api"

  //可省略，默认为`GET`
  methods = [ "GET" ]

  //参见【方法定义语法】

}

```

### 方法定义语法

```hcl

// method只能为：get、post、put、patch、delete
method {

  //服务接口的分类标签（用于swagger文档输出，可忽略），可继承自上级定义
  tags = ["标签"]

  //摘要描述（用于swagger文档输出，可忽略）
  desc = ""

  //预先执行其他接口：接口列表
  include = ["_boot"]

  //校验表达式：JS脚本实现
  validators {
    value = "express value" //表达式为真表示校验通过
  }

  //身份验证：返回true表示身份验证通过（可忽略）
  authorizer = <<JS
    (function(){
      user_name = $input.http_iv_user
      ...
      return true
    })()
  JS

  //安全验证配置：需要运行时配置统一安全验证服务地址
  //安全配置默认自上层定义中继承
  security {

    //是否允许匿名访问，为true时不判定用户身份
    anonymous = false

    //用户所属组织域：参见<https://snz1.cn/k8s/javadoc/sc-client-api/doc/org/scope.html>
    scope = "employee"

    //判定角色列表
    roles = [ "ADMIN" ]

    //判定用户列表
    users = [ "neeker" ]

    //角色或用户判定策略：
    //      为allow时表示请求用户必须包含roles中定义的角色、用户必须在users定义的列表中
    //      为deny时表示请求用户不能是roles定义的角色、用户不能再users定义的列表中
    //条件不满足则返回403应答
    policy = "allow"

  }

  //Redis缓存配置（无redis连接配置时无效）
  cache {
    //缓存名称列表(HSET)：使用input作为field主键
    put = ["cache_name"]
    //移除缓存(HSET)：删除HSET主键
    evit = ["cache_name"]
  }

  //表示total与exec的脚本类型：sql（默认），js（JavaScript），cmd（命令）
  impl = "sql"

  //SQL执行变量绑定，impl = "sql"时生效
  bind {
    sql_param1 = "$input.id"     //JS脚本
    sql_offset = "$input.offset" //默认为0
    sql_limit = "$input.limit"   //默认为0
  }

  //提供待执行的脚本（JS）：存在则忽略total、exec的定义
  provider = <<JS
  (function(){
    if ($result = "total") {
      return "..."
    } else {
      return "..."
    }
  })()
  JS

  //分页对象的总记录数，此属性存在则result类型强制为page
  total = <<SQL
    select count(*) from
  SQL

  //返回数据的执行SQL
  exec = <<SQL
    select * from your_table where id = :id offset :sql_offset limit :sql_limit
  SQL

  //转换为最终输出的JSON对象：JS语法，应当包含一个JS闭包函数
  transformer = <<JS
    (function(){
        //$result表示
        $new = [];
        for ( i in $result ) {
            $new.push($result[i].Database)
        }
        return $new
    })()
  JS

  //返回类型：object表示对象、list表示列表、page表示分页。
  result = "object"

  //应答格式，enclosed表示接口返回信封封装（默认），origin表示原样返回（不封装）
  //nil表示不返回数据（仅返回封装头），redirect表示跳转到返回地址
  format = "enclosed"

  //执行并组合其他接口返回值，存在则忽略其他定义
  aggregate = [ ... ]

  //定义Websocket接口
  websocket {

    //是否启用
    enabled = true

    //握手超时时间
    handshakeTimeout = 60

    //保持会话秒
    keepalive = 25

    //读取缓冲区大小
    readBufferSize = 512

    //写入缓冲区大小
    writeBufferSize = 512

    //是否默认启用压缩
    compression = true

    //子协议数组
    subprotocols = [ ... ]

    //允许的请求源正则表达式数组
    origins = [...]
  }

  //映射的静态目录
  dir = "/path/to/directory"

  //映射的静态文件
  file = "/path/to/file"
}

```

### 接口定义项说明

#### 接口名称

以英语字母开头的英文名称，后面可以是字母、数字或`_`，不能包含`:`、`/`、`\`、`-`等特殊字符。

#### `path`

表示接口路径，可以使用`:`开头作为路径变量，每个路径变量只能匹配一层路径，如：

```url
/users/:id/name
```

#### `tags`

文档归类标签列表，可以把多个不同的接口规定到同一个文档标签下展示，仅用于`SwaggerUI`。

#### `desc`

接口实现概述，仅用于`SwaggerUI`。

#### `const`

常量列表定义，此处定义的常量在`js`脚本可通过`$const.xxx`方式获取到。常量值可以是`js`表达式或`js`脚本。

#### `get`

仅在**路径定义**中有效，表示定义指定路径的`GET`方法实现。

#### `post`

仅在**路径定义**中有效，表示定义指定路径的`POST`方法实现。

#### `put`

仅在**路径定义**中有效，表示定义指定路径的`PUT`方法实现。

#### `patch`

仅在**路径定义**中有效，表示定义指定路径的`PATCH`方法实现。

#### `delete`

仅在**路径定义**中有效，表示定义指定路径的`DELETE`方法实现。


#### `include`

执行其他接口实现。

#### `validators`

验证表达式，每个定义的验证配置项必须返回`true`，否则返回422应答。

#### `authorizer`

使用`js`实现身份验证。

#### `security`

统一用户安全配置，包括可否匿名访问（`anonymous`）、定义组织域（`scope`）、定义角色列表（`roles`）、定义用户列表（`users`）、判定策略（`policy`）等配置。

 - `policy`为`allow`时表示允许定义的角色或用户访问；
 - `policy`为`deny`时表示不允许定义的角色或用户访问；

#### `cache`

接口缓存配置，包括设置缓存列表（`put`）或清理缓存列表（`evit`）。


#### `impl`

表示接口的`exec`与`total`实现脚本类型：

  - 为`sql`时表示SQL查询语句（默认）
  - 为`js`时表示`JavaScript`脚本
  - 为`cmd`时表示命令行及参数


#### `bind`

`impl`为`sql`时表示`exec`中的`SQL`查询命名绑定参数。

`bind`也为`SwaggerUI`提供输入参数的定义描述。

#### `total`

存在时表示接口为分页接口，并返回查询记录总数。

#### `exec`

服务接口实现脚本

#### `transformer`

用于转换`exec`执行返回的数据，采用`js`脚本实现。

#### `result`

定义数据类型，`object`表示对象、`list`表示列表、`page`表示分页（`total`存在时强制为`page`）。

#### `format`

定义应答格式，`enclosed`表示信封封装、`origin`表示原样返回、`nil`表示只返回封装头。

#### `aggregate`

组合其他接口定义执行返回。

#### `websocket`

定义接口是否为`websocket`服务，当启用了`websocket`服务时，接口定义的脚本在客户端消息到达时被执行。

#### `dir`

定义静态文件目录路由，存在时忽略其他配置，只返回目标文件夹中的静态文件。

#### `file`

定义静态文件路由，存在时忽略其他配置，只返回目标静态文件。

