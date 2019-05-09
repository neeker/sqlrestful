_meta {
  mq {

    driver = "amqp"

    uri = "amqp://10.158.3.23:32341/"

  }
}

_consumer {

  consume {
    queue = "app.sqlrestful.test.amqp"
  }

  impl = "js"

  exec = <<JS
  (function(){
    log($input)
  })()
  JS

}
