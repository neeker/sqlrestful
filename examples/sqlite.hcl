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

//数据集接口
dataxs {

  tags = [ "1.测试接口" ]

  path = "/dataxs"

  //执行前引入
  include = [ "_create_db" ]

  //分页查询
  get {

    summary = "分页查询数据"

    //根据条件返回执行脚本
    provider = <<JS
    (function(){
      var exec = "SELECT * FROM datax"
      var total = "SELECT count(*) FROM datax"
      if (typeof($input.express) != "undefined" && $input.express.length > 0) {
        exec += " WHERE NAME like :express "
        total += " WHERE NAME like :express "
      }
      exec += " ORDER BY NAME LIMIT :offset OFFSET :limit"
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

    summary = "添加数据"

    //输入参数
    bind {
      id = "$input.id"
      name = "$input.name"
      age = "$input.age"
      address = "$input.address"
    }

    //执行插入并返回结果
    exec = <<SQL
      REPLACE INTO datax(id, name, age, address) VALUES(:id, :name, :age, :address)
      
      ---

      SELECT * FROM datax WHERE id = :id
    SQL

    //返回类型
    result = "object"

  }

}

//数据实例接口
datax {

  tags = [ "1.测试接口" ]

  path = "/dataxs/:id"

  //执行前引入
  include = [ "_create_db" ]

  get {

    summary = "根据ID获取数据"

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

  }

  put {

    summary = "根据ID修改数据"

    //参数绑定
    bind {
      id = "$input.id"
      name = "$input.name"
      age = "$input.age"
      address = "$input.address"
    }

    exec = <<SQL
      UPDATE datax SET name = :name, age = :age, address = :address WHERE id = :id

      ---

      SELECT * FROM datax WHERE id = :id
    SQL

    result = "object"

  }

  delete {

    summary = "根据ID修改数据"

    //参数绑定
    bind {
      id = "$input.id"
    }

    //执行删除
    exec = <<SQL
      DELETE FROM datax WHERE id = :id
    SQL

    //不返回数据
    ret = "null"

  }


}