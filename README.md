# SQL转Restful接口服务

## 设计初衷

一直在使用`Java+SpringBoot`作为微服务生产力工具，大致上实现一个Restful微服务接口需要做以下相关工作：

  - ORM映射实现：通常使用MyBatis，需要写Pojo类、Mapper类和SQLProvider类。
  - 服务层实现：需要一个对象管理接口类与一个对象管理实现类。
  - 控制层实现：一个RestConfoller类并注解Rest方法再调用服务层实现。
  - Devops配置：Dockerfile、k8s部署描述文件等等。

这样一个`Restful`接口实现下来至少需要`5`个以上的类，大部分工作是在做转换、校验等语言相关的工作。

我们可以想想从SQL到Restful经历了多少层的实现，而大部分工作是毫无意义的规则代码。。

> 您可能会建议我使用spring-cloud云原生开发框架，但一样也少不上述过程。

因此一直以来，我一直想要有一个工具可以直接把SQL转成Restful微服务接口，它必须是云原生的开发方式：

  - 1、配置化实现：通过简单的配置边可以很方便的实现SQL转Restful接口；
  - 2、执行效率高：不能因为配置和转换减低运行效率；
  - 3、可容器化部署：能方便打包成Docker镜像并运行；
  - 4、多数据库支持：包括oracle、db2、mysql、postgres、hadoop等。

通过此工具可以快速对外提供Restful规范的数据微服务j接口，满足碎片化的数据服务需求应用场景的快速响应。

> 说干就干，于是找到了[sqler](https://github.com/alash3al/sqler)，但是[sqler](https://github.com/alash3al/sqler)仅支持REST而不支持Restful。<br>
>因此我在其基础之上实现了一个完整的SQL转Restful接口的服务工具，在兼容[sqler](https://github.com/alash3al/sqler)配置语法
>的同时进行了Restful配置扩展实现。

感谢开源！

## 基础概念

### 什么是HCL配置语言

最终选择HCL语言作为本工具服务配置语言，具体参见[HCL官网](https://github.com/hashicorp/hcl)。

HCL是大名鼎鼎的云基础架构自动化工具[hashicorp](https://www.hashicorp.com/)实现的配置语言，
它吸收了`JSON`与`YAML`及一些脚本语言的特性，自身兼容`JSON`语法：

 - 单行注释以#或开头//

 - 多行注释包含在/*和中*/。不允许嵌套块注释。多行注释（也称为块注释）在第一个*/找到时终止。

 - 属性值设置用key = value（空格忽略）表示。value可以是字符串，数字，布尔值，对象或列表。

 - 字符串必须用双引号，可以包含任何UTF-8字符。例："Hello, World"

 - 多行字符串从一行<<EOF的末尾开始，并EOF结束。可以使用任何文本代替EOF。例：

```
    <<SQL
    hello
    world
    SQL
```

  - 数字默认为10禁止，如果前缀为0x的数字，则将其视为十六进制。如果它以0为前缀，则将其视为八进制。数字可以是科学记数法：“1e10”。

  - 布尔值：true，false

  - 数组可以通过包装来制作[]。示例： ["foo", "bar", 42]。数组可以包含基础类型、其他数组和对象。作为替代方案，可以使用以下结构使用重复的块创建对象列表：


```
      service {
          key = "value"
      }
      service {
          key = "value"
      }
```

### SQLRestful的配置结构

**基本宏定义**

```hcl

//用于restful接口中的get、post、put、patch、delete等属性定义
macro_define {

  //引入其他宏定义
  include = ["_boot"]

  //返回值类型：list（列表，默认）、object（对象）、page（分页）
  type = "list"

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

  //SQL执行变量绑定
  bind {
    sql_param1 = "$input.id" //也可以使用JS转换
    sql_offset = "$input.offset" //默认为0
    sql_limit = "$input.limit"   //默认为0
  }

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

}

```

**Restful对象集接口定义**

```hcl

//接口定义名称，不能包含“\”、“/”、“:”等特殊字符
object_items {

  //接口地址，省略时使用接口定义名称作为接口地址
  path = "/path/of/object_items"

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

**Restful对象接口定义**

```hcl

//接口定义名称，不能包含“\”、“/”、“:”等特殊字符
object_items {

  //接口地址，路径ID使用“:”作为前缀
  path = "/path/of/object_items/:id"

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


