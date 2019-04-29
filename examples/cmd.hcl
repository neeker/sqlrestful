dir {

  bind {
    "" = "$input.dir"
  }

  impl = "cmd"

  exec = "dir"

}

docker_pull {

  path = "/docker/pull"

  bind {
    pull = "$input.img"
  }

  impl = "cmd"

  exec = "docker"

}
