# 使用`JWT`身份令牌请求其他接口

## 应用场景

如果您的`SQLRestful`实现需要通过[应用网关调用其他微服务接口](https://snz1.cn/k8s/javadoc/appgateway/2.%E7%94%A8%E6%88%B7%E6%89%8B%E5%86%8C/CallApi.html 点击鼠标右键打开新窗口)时，必然需要使用`JWT`规范的
身份令牌来完成调用。

## 准备`JWT`应用身份资料

通过[应用网关调用其他微服务接口](https://snz1.cn/k8s/javadoc/appgateway/2.%E7%94%A8%E6%88%B7%E6%89%8B%E5%86%8C/CallApi.html 点击鼠标右键打开新窗口)约定，生成`JWT`请求令牌必须遵循`JWT RS256`算法约定，需要准备证明应用身份的资料：

  - `RSA`私钥：对应公钥在已在网关端登记注册
  - `JWT`安全令牌：协商密钥，由网关随机生成

## 以`JWT`应用身份运行容器

准备`JWT`应用身份资料后，我们需要把`RSA`私钥文件加载到镜像的文件系统中，然后在运行命令行中加入`jwt-keyfile`、`jwt-secret`、`jwt-expires`参数，如下示例命令：

```sh
docker run -ti --rm \
  -v /path/of/app.pem:/sqlrestful/app.pem:ro \
  snz1/sqlrestful \
  -jwt-keyfile "/sqlrestful/app.pem" \
  -jwt-secret "***********" \
  -jwt-expires=3600 \
  ...
```

## 在文件中配置`JWT`应用身份

```hcl
_meta {

  //...

  jwt {
    //RSA私钥（PEM格式）
    rsa = <<PEM
    //PEM私钥
    PEM

    //协商密钥
    secret = "*****"

    //请求令牌超时时间（秒）
    expires = 1800
  }

  //...

}
```

## 使用`JWT`请求令牌调用其他接口

通过上述配置后在`JavaScript`脚本中我们便可以使用内置的`call_api`函数携带令牌请求其他微服务接口：

```js

(function(){
  var ratdata = call_api("http://appgateway.domain/paht/of/api", {
    method: "GET",
    headers: {
      ...
    },
    body: {
      ...
    }
  })
})()

```

> 内置的`call_api`函数根据`JWT`应用身份资料自动生成`JWT`请求令牌请求指定的接口。

## 直接获取`JWT`请求令牌

```js
(function(){
  var jwt_token =  jwt_token()
})()
```
