# 使用外部`SHELL`命令实现接口

## 应用场景

某些情况下可能需要使用已有SHELL脚本或已实现的命令行程序来快速对外提供`Restful`微服务接口：

  - 想要快速把命令行操作转成`Restful`微服务接口（自动化运维应用）

## 配置方法

在接口定义上使用`impl = "cmd"`配置后可直接对`exec`实现调用`SHELL`命令，如下所示：

```hcl
  path = "/docker_pull"
  bind {
    pull = "$input.img"
  }

  impl = "cmd"

  exec = "docker"
```

参数绑定规则：

在`bind`中配置的`SHELL`命令输入参数会以如下方式传入到`exec`配置的命令中：

```sh
docker pull <img>
```

> 命令执行展开规则：`<command> <bind name> <bind param value> ...`

## 命令执行返回

客户端请求命令实现的微服务接口时阻塞直至命令执行完成，`SQLRestful`获取`SHELL`命令的控制台输出字符串返回。

> 返回数据格式参见《[`Restful`接口返回的数据格式](ret.md)》章节。

