# 使用`JavaScript`实现接口

## 应用场景

默认情况下`SQLRestful`接口使用`SQL`来实现，但某些场景下我们不需要使用`SQL`而是需要调用其他`Restful`微服务接口：

  - 多个`Restful`微服务接口的数据组合；
  - 需要对已有的`Restful`数据做脱敏等实现；
  - 第三方公有云服务接口的封装等；

## 配置方法

在接口定义上使用`impl = "js"`配置后可直接对`exec`或`total`实现采用`JavaScript`，如下所示：

```hcl
  impl = "js"

  total = <<JS
  
  (function(){
    total = 0
    ...
    return total
  })()

  JS

  exec = <<JS
    var data
    ...
    return data
  JS
```

> 此处的`JavaScript`可以使用《[使用`JavaScript`脚本](js.md)》章节中说明的内置函数。
