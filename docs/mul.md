# 执行多段`SQL`

在某些情况下一个`exec`实现可能需要分为多个`SQL`段执行，在脚本加入“`---`”分隔符，如下代码所示：

```
exec = <<SQL

- 这里插入
insert into ...

---

- 这里是返回结果
select * from ....

SQL
```

> 注意： `exec`与`total`始终以最后一段`SQL`的返回作为返回数据。
