workflow "build_images" {
  on = "push"
  resolves = [
    "push pty-device-plugin",
    "push pty-client",
  ]
}

action "login @docker_hub" {
  uses = "actions/docker/login@master"
  secrets = [
    "DOCKER_PASSWORD",
    "DOCKER_USERNAME",
  ]
}

action "build pty-device-plugin" {
  uses = "actions/docker/cli@master"
  args = "build --build-arg TARGET=pty-device-plugin -t arhatdev/pty-device-plugin:latest -f cicd/docker/app.dockerfile ."
}

action "build pty-client" {
  uses = "actions/docker/cli@master"
  args = "build --build-arg TARGET=pty-client -t arhatdev/pty-client:latest -f cicd/docker/app.dockerfile ."
}

action "push pty-device-plugin" {
  needs = ["build pty-device-plugin", "login @docker_hub"]
  uses = "actions/docker/cli@master"
  args = "push arhatdev/pty-device-plugin:latest"
}

action "push pty-client" {
  needs = ["build pty-client", "login @docker_hub"]
  uses = "actions/docker/cli@master"
  args = "push arhatdev/pty-client:latest"
}
