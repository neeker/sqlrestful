# 使用用户统一安全认证服务

## 应用场景

如果我们的接口面向用户前端，此时必然涉及用户身份验证及权限验证问题，虽然通过`authorizer`配置项可以使用`JavaScript`脚本来实现用户身份验证及权限验证，但总的来说实现还是稍显繁琐了。

此时我们可以通过`security`配置项来对接口定义用户及权限角色访问控制。

> 注：security配置项需要[用户统一安全认证服务](https://snz1.cn/k8s/javadoc/sc-client-api/)组件支持。

## 配置方法

`security`可以针对接口配置以下验证方式：

  - 可以访问接口的登录用户或角色
  - 不能访问接口的登录用户或角色

```

  security {

    //是否允许匿名访问，为true时不判定用户身份
    anonymous = false

    //用户所属组织域：参见<https://snz1.cn/k8s/javadoc/sc-client-api/doc/org/scope.html>
    scope = "employee"

    //判定角色列表
    roles = [ "ADMIN" ]

    //判定用户列表
    users = [ "neeker" ]

    //角色或用户判定策略：
    //      为include时表示请求用户必须包含roles中定义的角色、用户必须在users定义的列表中
    //      为exclude时表示请求用户不能是roles定义的角色、用户不能再users定义的列表中
    //条件不满足则返回403应答
    policy = "allow"

  }

```

## 运行配置

开启`security`配置功能需要在启动命令行中加入`uumapi`、`useridtype`、`userscope`参数：

  - `uumapi`表示[用户统一安全认证服务](https://snz1.cn/k8s/javadoc/sc-client-api/)地址
  - `useridtype`表示通过请求头获取到的用户身份标识类型（包括：id、uname等）
  - `userscope`表示登录用户所在的组织域代码（登录的用户必须在该组织域下才能访问）
