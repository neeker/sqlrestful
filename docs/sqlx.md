# 根据请求参数实现条件分支

## 应用场景

有时我们需要根据不同的请求参数来组装`SQL`的不同条件语句，例如当请求参数传入了年龄参数就需要获取符合年龄条件的
人员，如果请求未传入年龄参数则不需要根据年龄过滤数据，此类场景在一条`SQL`查询语句中显然无法实现。

此时便需要我们使用执行提供器配置方式（`provider`配置）来完成复杂的`SQL`条件分支实现。

## `provider`

`provider`配置采用`JavaScript`实现，定义了`provider`意味着微服务接口接下来的执行脚本（包括`total`与`exec`)由`provider`定义的`JavaScript`返回。

### 只返回`exec`执行脚本

`provider`返回格式可以是简单的字符串，返回的字符串直接用于`exec`执行，如下所示：

```hcl
  provider = <<SQL
  (function(){
    return "SELECT * FROM ..."
  })()
  SQL
```

### 同时返回`total`与`exec`

`provider`可以返回用于分页查询接口的`total`与`exec`脚本，如下所示：

```hcl
  provider = <<SQL
  (function(){

      return {
        total: "SELECT count(*) FROM ...",
        exec:  "SELECT * FROM ... OFFSET :offset LIMIT :limit"
      }


  })()
  SQL
```

### 同时返回脚本类型


```hcl
  provider = <<SQL
  (function(){

      return {
        total: "SELECT count(*) FROM ...",
        exec:  "SELECT * FROM ... OFFSET :offset LIMIT :limit",
        impl:  "sql"
      }


  })()
  SQL
```

- `total`可以不存在于返回结果中，存在则表示实现分页查询接口
- `impl`可以是`js`（表示`JavaScript`脚本）、`cmd`（表示SHELL脚本命令）。

