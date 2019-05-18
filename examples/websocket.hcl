
_meta {
  name = "WebSocket服务"
  desc = "WebSocket服务实例"
}

root {
  
  path = "/"

  impl = "js"

  exec = "'/websocket.html'"

  format = "redirect"

}

static {
  path = "/websocket.html"

  file = "examples/websocket.html"
}

echo {
  path = "/echo"

  impl = "js"

  websocket {
    enabled = true
  }

  exec =<<JS
  (function(){
    log($input)
    return "ok"
  })()
  JS

}

