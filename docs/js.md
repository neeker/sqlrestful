# `SQLRestful`的`JS`脚本能力

## `js`脚本用途

在SQLRestful的主要实现由SQL与JavaScript完成，其中JavaScript负责完成下述工作：

  - 与其他微服务接口进行交互；
  - 实现请求身份验证功能；
  - 校验请求参数是否合法；
  - 请求参数到SQL绑定参数转换；
  - SQL返回结果的转换能力；
  - 实现微服务逻辑处理功能；

它支持完整的`ECMAScript 5.1`规范（由 [goja](https://github.com/dop251/goja) 提供实现支持）。

## `js`脚本的默认变量

在SQLRestful的宏定义中，参数转换（`bind`宏），身份验证实现（`authorizer`宏）的`JS`脚本可以通过`$input`变量
获取到请求输入参数列表，同时可以通过`$input.http_xxxx`来获取请求头内容，假设客户端请求发送一个名称为`X-Test-MM`
的请求，则通过以下表达式拿到请求头内容：

```
$input.http_x_test_mm
```

> 所有请求头都会转换成小写加上`http_`前缀，同时为了遵循`JS`对象属性命名规则会把“`-`”替换为“`_`”。

应答转换（`transformer`宏）的脚本通过变量`$result`可以获取到`exec`宏返回的原始数据，具体见如下示例代码：

```
test {

  ...

  transformer = <<JS
    (function(){
    $new_result = $result
    $new_result.trans_test = "13456"
    return $new_result  
    })()
  JS

  ...

}
```

除此之外，每个`JS`脚本都可以获取以下参数：

| 参数名 | 参数说明 |
| ------ | -------------------------------------- |
| `$input.http_method` | 当前请求方法：GET、POST、PUT、PATCH、DELETE |
| `$input.http_scheme` | 当前请求协议头：http、https |
| `$input.http_path` | 当前请求上下文路径 |
| `$input.http_url` | 当前请求URL地址 |
| `$input.http_uri` | 当前请求URI地址 |
| `$input.http_remote_addr` | 当前请求远程IP地址（或前置代理地址） |
| `$input.http_real_ip` | 当前请求真实IP地址：<br>请求头`X-Forwarded-For`或`X-Real-IP`的值 |
| $name | 当前微服务实现宏名称 |
| $stage | 宏`JS`所在过程名称：validators、authorizer、<br>transformer、bind、provider、exec、total |


## `js`脚本内置函数


`SQLRestful`为`JS`脚本内置了两个默认的`HTTP`请求函数和一个控制台日志输出函数：

  - `fetch`
  - `call_api`
  - `log`

### `fetch`函数

**函数原型**

```
function fetch(URL, {
  method: "HTTP METHOD", //请求方法，如GET、POST、PUT
  headers: { //请求头
    ...
  },
  body: ... //请求体，可以是JSON或字符串。
})
```

**返回结果**

```
{
  "status":     "应答状态文本",
  "statusCode": "HTTP应答码",
  "headers":    "HTTP应答头",
  "body":       "应答内容字符串",
}
```


### `call_api`函数

此函数提供后台调用在[应用网关](https://snz1.cn/k8s/javadoc/appgateway/2.%E7%94%A8%E6%88%B7%E6%89%8B%E5%86%8C/ExpSvc.html)中注册的微服务[后台接口](https://snz1.cn/k8s/javadoc/appgateway/2.用户手册/ExpSvc.html#认证模式说明)。

它通过SQLRestful服务配置的 JWT RSA 私钥与 JWT 安全令牌产生 JWT 请求令牌并发起接口请求。

> 具体JWT令牌请求方式参见《[通过网关调用后台服务接口
](https://snz1.cn/k8s/javadoc/appgateway/2.%E7%94%A8%E6%88%B7%E6%89%8B%E5%86%8C/CallApi.html)》中的说明。

**函数原型**

```
function call_api(URL, {
  method: "HTTP METHOD", //请求方法，如GET、POST、PUT
  headers: { //请求头
    ...
  },
  body: ... //请求体，可以是JSON或参数内容。
})
```

**返回结果**

正常情况下 call_api 函数直接返回接口的JSON对象，只有在请求出错的情况下返回如下定义：


```
{
  "status":     "应答状态文本",
  "statusCode": "HTTP应答码",
  "headers":    "HTTP应答头",
  "body":       "应答内容字符串",
}
```

如果请求的接口应答内容不能转换为JSON对象则返回与`fetch`函数相同的应答：

### `log`函数


**函数原型**

```
function log(message, ...)
```

**使用示例**

```js
(function(){
  log("debug:", "hello world!")
})()
```
