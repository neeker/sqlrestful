# `SQL`实现`Restful`接口说明

`SQLRestful`默认采用`SQL`来实现`Restful`接口（`total`与`exec`配置定义）。

## SQL参数绑定

要求通过`bind`宏来配置`SQL`可变参数。其中的属性定义值是`JavaScript`表达式（也可以是`JavaScript`闭包函数），参见如下示例配置：

```hcl
bind {
  name_arg = "'%' + $input.name + '%'"
  end_arg = <<JS
    (function(){
      return '%' + $input.name
    })()
  JS
}
```

> `$input`表示接口请求参数对象集，具体参见下一章《[`SQLRestful`的`JS`脚本能力](js.md)》


通过上述配置后，可以在`exec`或`total`的`SQL`代码中使用“:<变量名>”的绑定参数，如下所示：

```hcl
exec = <<SQL
  select * from tbname where name like :name_arg or name like :end_arg
SQL
```

## SQL分段执行方式

在某些情况下一个`exec`实现可能需要分为多个`SQL`段执行，在脚本独立的一行中加入“`---`”分隔符，如下代码所示：

```
exec = <<SQL

- 这里是入口
insert into ...

---

- 这里是返回结果
select * from ....

SQL
```

> 注意： `exec`与`total`始终以最后一段`SQL`的返回作为返回数据。

