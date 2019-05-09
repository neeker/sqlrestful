_meta {
  mq {

    driver = "stomp"

    uri = "tcp://user:user9527@test.snz1.cn:61613/"

  }

}

_consumer {

  consume {
    name = "app.sqlrestful.test.stomp"
  }

  impl = "js"

  exec = <<JS
  (function(){
    log($input)
  })()
  JS

}
