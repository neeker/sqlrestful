//对象集
tables {

  //上下文路径
  path = "/tables"

  //获取对象集数据
  get {

    //
    tag = ["测试接口"]

    //实现概要
    summary = "获取数据库中的数据表"

    //返回记录总数，加了此定义则返回分页对象(强制type=page)
    total = <<SQL
      SELECT count(tablename) FROM pg_tables 
      WHERE tablename NOT LIKE 'pg%' AND tablename NOT LIKE 'sql_%';
    SQL

    //请求验证
    authorizer = <<JS
    (function(){
      
      //获取请求头“X-Credential-Userid”内容，通过网关请求的用户身份ID
      var user_id = $input.http_x_credential_userid
      if (user_id === undefined || user_id === "") {
        //return false
      }
      return true
    })()
    JS

    //输入参数绑定
    bind {
      offset = "$input.offset"
      limit = "$input.limit"
    }

    //缓存
    cache {
      //返回并设置缓存
      put = ["test.tables"]
    }

    //接口返回SQL表达式
    exec = <<SQL
      SELECT * FROM pg_tables 
      WHERE tablename NOT LIKE 'pg%' AND tablename NOT LIKE 'sql_%' 
      ORDER BY tablename  offset :offset limit :limit;
    SQL

  }

}

//对象
table_item {

  path = "/tables/:id"

  get {

    //返回对象类型：object表示单个对象
    type = "object"

    //参数绑定，input表示请求参数
    bind {
      tablename = "$input.id"
    }

    //缓存配置
    cache {
      //返回并设置缓存
      put = ["test.table"]
    }

    //接口返回SQL表达式
    exec = <<SQL
      SELECT * FROM pg_tables 
      WHERE tablename = :tablename
    SQL

    //返回转换
    transformer = <<JS
    (function(){
      // $result 为函数输入参数
      $new_result = $result;
      response = call_api("http://test.snz1.cn:8090/xeai/users/admin", {
        method: 'GET'
      })
      if (response.code != 0) {
        throw response.message
      } else {
        $new_result.user = response.data
        $new_result.cache_date = new Date();
      }
      return $new_result
    })()
    JS

  }

}

//JS实现接口
test_js {

  //接口地址
  path = "/users/:uid"

  get {
    //使用JavaScript实现
    impl = "js"

    //实现脚本
    exec = <<JS
    (function(){
      // $input 为函数输入参数
      return call_api("http://test.snz1.cn:8090/xeai/users/" + $input.uid, {
        method: 'GET'
      })
    })()
    JS

    //不封装返回类型
    ret = "origin"
  }

}
