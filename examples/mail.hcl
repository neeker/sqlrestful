
_meta {

  smtp {

    host = "localhost"

    port = 25

    username = "test"

    password = "test"
    
    ssl = false

    address = "mail"

    name = "name"

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
      try {
        send_mail($input.to, $input.cc, $input.bcc, $input.subject, $input.body)
      } catch(err) {
        throw err.message
      }
      return true
    })()
    JS

  }

}