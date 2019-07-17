test {
  timer {
    inteval = 3
  }
  impl = "js"

  exec =<<JS
  (function(){
    log("test" + new Date())
  })()
  JS
}
