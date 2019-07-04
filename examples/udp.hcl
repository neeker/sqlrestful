
_udp {

  udp {
    port = 9182
  }

  impl = "js"

  exec = <<JS
  (function(){
    log($input.data)
  })()
  JS

}

sendudp {

  impl = "js"

  exec =<<JS
  
  (function(){
    send_udp("127.0.0.1:9182", "testdata")
  })()

  JS

}
