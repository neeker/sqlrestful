_meta {
  mq {

    driver = "stomp"

    uri = "tcp://user:user9527@test.snz1.cn:61613/"

  }

}

_consumer {

  consume {
    // topicq前必须加/topic/，如：/topic/app.foo.bar
    name = "app.foo.bar"
  }

  impl = "js"

  exec = <<JS
  (function(){
    log($input)
  })()
  JS

}
