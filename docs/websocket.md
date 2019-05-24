# 实现websocket服务接口

## 应用场景

当我们需要与客户端建立长连接通道时需要使用`websocket`。

SQLRestful中定义的`websocket`服务接口采用客户端消息驱动机制实现：

每次任何一个客户端消息到达时都会执行接口定义的`exec`脚本，其会把最终执行结果作为一个JSON消息对象发送给消息发送客户端（定义`format`等于`nil`时则不返回消息）


## 实现`websocket`接口

### 语法结构

```
test {
  websocket {
    enabled = true
  }

  impl = "js"

  exec = <<JS
  ...
  JS  
}

```

在`exec`中可以通过以下变量获取请求信息：

  - `$input.__clientid__`： 获取当前发送消息的客户端ID
  - `$input.data`：获取到客户端发送的消息对象

### 可用的`js`函数

#### `ws_broacast`

```js
ws_broacast('<接口名称，如：test>', {
  ...
})
```

#### `ws_send`

```js
ws_send('接口名称，如：test', '客户端ID', {
  ...
})
```

### 其他说明

通过消息队列与数据库配合即可复杂的`websocket`应用。

