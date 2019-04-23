//对象集
tables {

  //上下文路径
  path = "/tables"

  //获取对象集数据
  get {

    //返回记录总数，加了此定义则返回分页对象(强制type=page)
    total = <<SQL
      SELECT count(tablename) FROM pg_tables 
      WHERE tablename NOT LIKE 'pg%' AND tablename NOT LIKE 'sql_%';
    SQL

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
      //清除缓存test.tables
      evit = ["test.tables"]
    }

    //接口返回SQL表达式
    exec = <<SQL
      SELECT * FROM pg_tables 
      WHERE tablename = :tablename
    SQL
  }

}
