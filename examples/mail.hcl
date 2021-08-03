
_meta {

  smtp {

    host = "localhost"

    port = 25

    username = "test"

    password = "test"
    
    ssl = false

  }

}

//发送邮件
sendmail {

  get {

    //概要描述
    desc = "发送邮件"

    impl = "js"

    // 直接发送邮件
    exec = <<JS
    (function(){
      var fromName = $input.fromName
      if (fromName == undefined) {
        fromName = ""
      }
      try {
        send_mail($input.from, fromName, $input.to, $input.subject, $input.body)
      } catch(err) {
        throw err.message
      }
      return true
    })()
    JS

  }

}