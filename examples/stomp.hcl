_meta {
  mq {

    driver = "stomp"

    url = "tcp://user:user9527@test.snz1.cn:61613/"

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
    var msg = JSON.stringify($input)
    log(msg)
    log(emit_msg('app.foo.bar', msg))
  })()
  JS

}
