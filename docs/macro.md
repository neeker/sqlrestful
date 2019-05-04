# `Restful`接口定义语法结构

## 接口宏定义语法结构

```hcl

//用于restful接口中的get、post、put、patch、delete等属性定义
api_name {

  //服务接口的分类标签（用于swagger文档输出，可忽略），可继承自上级定义
  tags = ["标签"]

  //摘要描述（用于swagger文档输出，可忽略）
  desc = ""

  //引入其他宏定义
  include = ["_boot"]

  //返回值类型：list（列表，默认）、object（对象）、page（分页）
  result = "list"

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


  //表示total与exec的脚本类型：sql（默认），js，exec（命令）
  impl = "sql"

  //SQL执行变量绑定，impl = "sql"时生效
  bind {
    sql_param1 = "$input.id"     //JS脚本
    sql_offset = "$input.offset" //默认为0
    sql_limit = "$input.limit"   //默认为0
  }

  //提供待执行的脚本（JS）
  provider = <<JS
  (function(){
    if ($result = "total") {
      return "..."
    } else {
      return "..."
    }
  })()
  JS

  //分页对象的总记录数，此属性存在则type类型强制为page
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
  ret = "enclosed"

  //组合其他接口返回值，存在则忽略其他定义
  aggregate = [ ... ]

}

```

## `Restful`接口语法结构

```hcl

//接口定义名称，不能包含“\”、“/”、“:”等特殊字符
name {

  //接口地址，省略时使用接口定义名称作为接口地址，可使用“:变量”方式定义路径变量
  path = "/path/of/object_items"

  //服务接口的分类标签（可忽略）
  tags = ["标签"]

  //摘要描述（可忽略）
  desc = ""

  //接口GET请求方法宏定义
  get {
   ... //基本宏定义
  }

  //POST请求方法宏定义
  post {
    ... //基本宏定义
  }

  //PUT请求方法宏定义
  put {
    ... //基本宏定义
  }

  //PATCH请求方法宏定义
  patch {
    ... //基本宏定义
  }

  //DELETE请求方法宏定义
  delete {
    ... //基本宏定义
  }

  //以下配置属性可以与【接口宏定义】一致，用于简化配置只有一种操作的接口定义
  ...

}

```
