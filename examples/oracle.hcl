test {

  tags = ["1.Oracle数据库测试"]

  path = "/tables"

  get {

    desc = "获取数据库中所有表"
    
    exec = <<SQL

    select t.table_name from user_tables t

    SQL

    model {
      TABLE_NAME {
        type = "string"
      }
    }

  }

}
