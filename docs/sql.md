# `SQL`参数绑定

`SQLRestful`默认采用`SQL`来实现`Restful`接口（参见`total`与`exec`定义项）。

我们通过`bind`来配置`SQL`动态命名参数。其中的属性定义值是`JavaScript`表达式（也可以是`JavaScript`闭包函数），参见如下示例配置：

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

> `$input`表示接口请求参数对象集，具体请参见《[使用`JavaScript`脚本](js.md)》章节。


通过上述配置后，可以在`exec`或`total`的`SQL`代码中使用“:<变量名>”的绑定参数，如下所示：

```hcl
exec = <<SQL
  select * from tbname where name like :name_arg or name like :end_arg
SQL
```
