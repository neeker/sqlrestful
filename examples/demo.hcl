
//自动建表
_create_db {

  exec = <<SQL
    CREATE TABLE IF NOT EXISTS test (
        name    TEXT  PRIMARY KEY  NOT NULL,
        hello          TEXT    NOT NULL
    );
  SQL

}

//测试接口
test {

  //定义接口标签
  tags = ["1.示例接口"]

  //引入自动建表
  include = ["_create_db"]

  //定义接口地址
  path = "/test"

  //获取分页数据接口
  get {

    //概要描述
    summary = "获取分页数据"

    //绑定参数
    bind {
      offset = "$input.offset"
      limit = "$input.limit"
    }

    //返回记录总数
    total = <<SQL
      SELECT COUNT(*) FROM test;
    SQL

    //返回分页数据
    exec = <<SQL
      SELECT name, hello FROM test LIMIT :limit OFFSET :offset;
    SQL

  }

  //新增数据
  post {

    //概要描述
    summary = "新增数据"

    //参数绑定
    bind {
      name = "$input.name"
      hello = "$input.hello"
    }

    //插入数据库并返回插入记录
    exec = <<SQL
  
      INSERT OR REPLACE INTO test (name, hello) VALUES(:name, :hello);

      ---

      SELECT name, hello FROM test WHERE name = :name;

    SQL

    //返回对象
    result = "object"

  }

}

