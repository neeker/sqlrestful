
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
    keepalive = 180
  }

  exec =<<JS
  (function(){
    ws_send('echo', $input.__clientid__, {
      "your": $input.__clientid__,
      "echo": $input.data
    })
    ws_broacast('echo', {
      "from": $input.__clientid__,
      "echo": $input.data
    })
    return "ok"
  })()
  JS

}

