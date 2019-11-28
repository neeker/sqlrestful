//描述信息
_meta {

  //名称
  name = "SQLRestful"

  //版本
  version = "1.0"

  //描述
  desc = "SQLRestful，您的云原生应用生产力工具！"

  //作者
  author {
    //姓名
    name = "痞子飞猪"
    //邮箱
    email  = "13317312768@qq.com"
  }

  //数据库配置
  db {
    //数据库驱动名
    driver = "sqlite3"
    dsn = "sqlte3.db?create=true"
  }

}

//SQL数据库
_create_db {

  //创建表
  exec = <<SQL

  CREATE TABLE IF NOT EXISTS datax (
      id INT PRIMARY KEY     NOT NULL,
      name           TEXT    NOT NULL,
      age            INT     NOT NULL,
      address        TEXT
  );

  SQL

}

root {

  path = "/"

  desc = "跳转至接口文档"

  tags = [ "1.SQL测试接口" ]

  get {
    impl = "js"

    exec = <<JS
    (function(){
      return "/doc.html"
    })()
    JS

    format = "redirect"
  }
}

//请求其他接口

includes {

  tags = [ "1.SQL测试接口" ]

  desc = "JS执行其他返回"


  //SQL参数绑定
  bind {
    offset = "$input.offset"
    limit = "$input.limit"
    express = "'%' + $input.express + '%'"
  }

  impl = "js"

  exec = <<JS
  (function(){
    return call_macro("dataxs.get", {
      offset: $input.offset,
      limit: $input.limit,
      express: $input.express == undefined ? "" : $input.express,
    })
  })()
  JS

  format = "origin"

}

//数据集接口
dataxs {

  tags = [ "1.SQL测试接口" ]

  path = "/dataxs"

  //执行前引入
  include = [ "_create_db" ]

  //分页查询
  get {

    desc = "分页查询数据"

    //根据条件返回执行脚本
    provider = <<JS
    (function(){
      var exec = "SELECT * FROM datax"
      var total = "SELECT count(*) FROM datax"
      if (typeof($input.express) != "undefined" && $input.express.length > 0) {
        exec += " WHERE NAME like :express "
        total += " WHERE NAME like :express "
      }
      exec += " ORDER BY NAME LIMIT :limit OFFSET :offset"
      return {
        total: total,
        exec:  exec
      }
    })()
    JS

    //SQL参数绑定
    bind {
      offset = "$input.offset"
      limit = "$input.limit"
      express = "'%' + $input.express + '%'"
    }

  }

  //添加数据
  post {

    desc = "添加数据"

    //输入参数
    bind {
      id = "$input.id"
      name = "$input.name"
      age = "$input.age"
      address = "$input.address"
    }

    //执行插入并返回结果
    exec = <<SQL
      REPLACE INTO datax(id, name, age, address) VALUES(:id, :name, :age, :address);
      
      ---

      SELECT * FROM datax WHERE id = :id;
    SQL

    //应答格式
    model {
      id {
        type = "integer"
      }
      name {
        type = "string"
      }
      age {
        type = "integer"
      }
      address {
        type = "string"
      }
    }

    //返回类型
    result = "object"

  }

}

//数据实例接口
datax {

  tags = [ "1.SQL测试接口" ]

  path = "/dataxs/:id"

  //执行前引入
  include = [ "_create_db" ]

  get {

    desc = "根据ID获取数据"

    //参数绑定
    bind {
      id = "$input.id"
    }

    //实现脚本
    exec = <<SQL
    SELECT * FROM datax WHERE id = :id;
    SQL

    //返回类型
    result = "object"

    //应答格式
    model {
      id {
        type = "integer"
      }
      name {
        type = "string"
      }
      age {
        type = "integer"
      }
      address {
        type = "string"
      }
    }

  }

  put {

    desc = "根据ID修改数据"

    //参数绑定
    bind {
      id = "$input.id"
      name = "$input.name"
      age = "$input.age"
      address = "$input.address"
    }

    exec = <<SQL
      UPDATE datax SET name = :name, age = :age, address = :address WHERE id = :id;

      ---

      SELECT * FROM datax WHERE id = :id;
    SQL

    result = "object"

    //应答格式
    model {
      id {
        type = "integer"
      }
      name {
        type = "string"
      }
      age {
        type = "integer"
      }
      address {
        type = "string"
      }
    }

  }

  delete {

    desc = "根据ID修改数据"

    //参数绑定
    bind {
      id = "$input.id"
    }

    //执行删除
    exec = <<SQL
      DELETE FROM datax WHERE id = :id;
    SQL

    //不返回数据
    format = "nil"

  }


}

jssqltest {

  tags = [ "2.JS测试接口" ]

  //执行前引入
  include = [ "_create_db" ]

  desc = "JS执行SQL查询"

  path = "/jssqltest"

  impl = "js"

  //这里的定义主要是为了生成SwaggerUI参数
  bind = {
    id = "数据ID"
  }

  exec = <<JS

  (function(){
    var ret = exec_sql("SELECT * FROM datax WHERE id = :id", {
      "id": $input.id
    })
    return ret === undefined ? undefined : ret[0]
  })()

  JS

}


jscmdtest {

  tags = [ "2.JS测试接口" ]

  desc = "JS执行命令"

  path = "/jscmdtest"

  impl = "js"

  exec = <<JS

  (function(){
    return exec_cmd("ls", "-l", "/")
  })()

  JS

}
