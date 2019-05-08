# 基本语法结构

`SQLRestful`语法中微服务接口定义分为两层结构：**路径定义**与**方法定义**。

## 路径定义语法

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

## 方法定义语法

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

  //返回数据处理，enclosed表示接口返回信封封装（默认），origin表示原样返回（不封装）
  //nil表示不返回数据（仅返回封装头），redirect表示跳转到返回地址
  format = "enclosed"

  //执行并组合其他接口返回值，存在则忽略其他定义
  aggregate = [ ... ]

}

```

## 定义项说明

### 接口名称

以英语字母开头的英文名称，后面可以是字母、数字或`_`，不能包含`:`、`/`、`\`、`-`等特殊字符。

### `path`

表示接口路径，可以使用`:`开头作为路径变量，每个路径变量只能匹配一层路径，如：

```url
/users/:id/name
```

### `tags`

文档归类标签列表，可以把多个不同的接口规定到同一个文档标签下展示，仅用于`SwaggerUI`。

### `desc`

接口实现概述，仅用于`SwaggerUI`。

### `const`

常量列表定义，此处定义的常量在`js`脚本可通过`$const.xxx`方式获取到。常量值可以是`js`表达式或`js`脚本。

### `get`

仅在**路径定义**中有效，表示定义指定路径的`GET`方法实现。

### `post`

仅在**路径定义**中有效，表示定义指定路径的`POST`方法实现。

### `put`

仅在**路径定义**中有效，表示定义指定路径的`PUT`方法实现。

### `patch`

仅在**路径定义**中有效，表示定义指定路径的`PATCH`方法实现。

### `delete`

仅在**路径定义**中有效，表示定义指定路径的`DELETE`方法实现。


### `include`

引用执行其他接口实现，