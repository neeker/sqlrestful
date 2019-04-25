
### JavaScript脚本

在SQLRestful的主要实现由SQL与JavaScript完成，其中JavaScript负责提供与其他微服务接口的交互、SQL返回结果的转换能力。

JavaScript主要用于参数转换（`bind`宏），身份验证实现（`authorizer`宏），应答转换（`transformer`宏）实现，它支持
完整的 ECMAScript 5.1 规范（由 [goja](https://github.com/dop251/goja) 提供实现支持）。

参数转换（`bind`宏），身份验证实现（`authorizer`宏）的JS脚本可以通过变量`$input`可以获取到请求输入参数：

* `$result`表示请求参数JSON对象
* 请求头中的参数通过`http_`开头+头名称（全部小写，`-`被替换成`_`），如有一个请求头叫`x-test-mm`，则通过以下表达式拿到值：

```
$input.http_x_test_mm
```

应答转换（`transformer`宏）的脚本通过变量`$result`可以获取到`exec`宏返回的原始应答JSON对象：

```
transformer = <<JS
(function(){
$new_result = $result
$new_result.trans_test = "13456"
return $new_result  
})()
JS
```

SQLRestful为JS脚本内置了两个默认的HTTP请求函数和一个控制台日志输出函数：

  - fetch
  - call_api
  - log

#### fetch 函数说明

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


#### call_api 函数说明

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
