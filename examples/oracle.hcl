test {

  path = "/tables"

  get {
    
    exec = <<SQL

    select t.table_name from user_tables t

    SQL

  }

}
