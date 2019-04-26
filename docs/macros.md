# `SQLRestful`开发及`HCL`配置说明


### 概要说明

SQLRestful采用HCL语言配合SQL、JavaScript脚本来开发微服务接口。

> 参见[什么是`HCL`配置语言](hcl.md)了解`HCL`基本概念

### 图解SQlRestful配置项

SQLRestful配置遵循Restful规范，采用路径、方法对应对象的不同操作语义，见下图所示：

<img src="img/sqlrestful.png" width="90%" />

#### 基本宏定义

```hcl

//用于restful接口中的get、post、put、patch、delete等属性定义
macro_define {

  //服务接口的分类标签（可忽略），可继承根定义
  tags = ["标签"]

  //摘要描述（可忽略）
  summary = ""

  //引入其他宏定义
  include = ["_boot"]

  //返回值类型：list（列表，默认）、object（对象）、page（分页）
  result = "list"

  //校验表达式：参见<https://github.com/asaskevich/govalidator>
  validators {
    value = "express value" //表达式为真表示校验通过
  }

  //身份验证：返回true表示身份验证通过（可忽略）
  authorizer = <<JS
    (function(){
      user_name = $input.iv_user
      ...
      return true
    })()
  JS

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
  //null表示不返回数据（仅返回封装头），redirect表示跳转到返回地址
  ret = "enclosed"

}

```

#### Restful对象集接口定义

```hcl

//接口定义名称，不能包含“\”、“/”、“:”等特殊字符
object_items {

  //接口地址，省略时使用接口定义名称作为接口地址
  path = "/path/of/object_items"

  //服务接口的分类标签（可忽略）
  tags = ["标签"]

  //摘要描述（可忽略）
  summary = ""

  //GET请求方法宏定义
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

  //以下属性兼容sqler

  //SQL执行变量绑定
  bind {
    sql_param1 = "$input.param1"
  }

  //兼容sqler语法，存在则忽略get、post、put、patch、delete的定义
  exec = <<SQL
    select * from your_table where id = :id offset :sql_offset limit :sql_limit
  SQL

  //兼容sqler语法
  methods = ["get"]
  
  //兼容sqler语法
  aggregate = ["macro_name"]

}

```

#### Restful对象接口定义

```hcl

//接口定义名称，不能包含“\”、“/”、“:”等特殊字符
object_items {

  //接口地址，路径ID使用“:”作为前缀
  path = "/path/of/object_items/:id"

  //服务接口的分类标签（可忽略）
  tags = ["标签"]

  //摘要描述（可忽略）
  summary = ""

  //GET请求方法宏定义
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

}

```

** 数据应答格式 **

默认情况接口应答采用信奉封装的JSON数据格式，基本格式如下：

```
{
  "code": 0,
  "message": "操作成功",
  "data": { //宏执行返回
    ...
  }
}
```

分页数据返回如下：

```
{
  "code": 0,
  "message": "操作成功",
  "data": { //宏执行返回
    "offset": 0, //起始索引
    "total": 0, //总记录数
    "data": [...] //分页数据列表
  }
}
```

> 你可以在接口配置上添加'ret="origin"'来禁用返回封装。

#### SQL参数绑定说明

如果使用`sql`作为接口实现语言时，可以通过`bind`宏来配置SQL参数。其中的属性定义值则是JS表达式（也可以是JS表达式）。

最终在SQL中可以使用“:变量名”的绑定参数，如下所示：

```
bind {
  name_arg = "'%' + $input.name + '%'"
  end_arg = <<JS
    (function(){
      return '%' + $input.name
    })()
  JS
}

exec = <<SQL
  select * from tbname where name like :name_arg or name like :end_arg
SQL
```

> `$input`表示请求参数。

