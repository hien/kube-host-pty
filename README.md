# kube-host-pty

[![Build Status](https://travis-ci.com/arhat-dev/kube-host-pty.svg)](https://travis-ci.com/arhat-dev/kube-host-pty) [![GoDoc](https://godoc.org/arhat.dev/kube-host-pty?status.svg)](https://godoc.org/arhat.dev/kube-host-pty) [![GoReportCard](https://goreportcard.com/badge/arhat-dev/kube-host-pty)](https://goreportcard.com/report/arhat.dev/kube-host-pty) [![codecov](https://codecov.io/gh/arhat-dev/kube-host-pty/branch/master/graph/badge.svg)](https://codecov.io/gh/arhat-dev/kube-host-pty)

a simple kubernetes device-plugin and toolset to access host pty

__WARNING: THIS PROJECT WORKS FOR ME, BUT NOT TESTED, USE AT YOUR OWN RISK__

## Purpose

- Learn `Kubernetes` device plugin by build one
- Eliminate unnecessary `ssh` identity management for my homelab `Kubernetes` cluster
- Enable Role-Based Access Control (RBAC) for host access with `Kubernetes`

## Components

- [pty-device-plugin](./cmd/pty-device-plugin) - A simple `Kubernetes device-plugin` deployed to host to expose host pty
- [pty-client](./cmd/pty-client) - A client to access pty exposed by `pty-device-plugin`
- (WIP) [kubectl-pty](./cmd/kubectl-pty) - A `kubectl` plugin to ease management of host pty served by `pty-device-plugin`

## Usage

1. Build and deploy `pty-devcie-plugin` to your `Kubernetes` node (`Golang` installation required), currently I don't provide esay package installation since this hasn't been tested

   ```bash
   # with GOPATH configured you can just go get on your `Kubernetes` node
   # and pty-device-plugin will be installed
   # $ go get -u arhat.dev/kube-host-pty/cmd/pty-device-plugin

   # or you need to download via git clone and build
   $ git clone https://arhat.dev/kube-host-pty
   $ cd kube-host-pty
   # build pty-device-plugin with make
   # wait to finish and you can find built pty-device-plugin at `./build/pty-device-plugin`
   $ make pty-device-plugin

   # deploy to your `Kubernetes` nodes and run with root, 
   #
   # instructions below are just for example
   # (local)   $ scp ./build/pty-device-plugin user@node.addr:~
   # (on node) $ sudo /path/to/pty-device-plugin --log=debug
   ```

2. Deploy `pty-client` with resource requests/limits `arhat.dev/pty` to those nodes when needed, here is a sample deployment script

   ```yaml
    apiVersion: v1
    kind: Pod
    metadata:
    name: pty-client-at-my-kube-node
    labels:
        app.kubernetes.io/name: pty-client
    spec:
    # use either `nodeAffinity`, `nodeSelector`, `toleration` as you wish
    nodeSelector:
        kubernetes.io/hostname: my-kube-node
    containers:
    - name: pty-client
        image: arhatdev/pty-client:latest
        command:
        - /app
        - --log=fatal
        # `stdin` and `tty` must be true to get client running properly
        stdin: true
        tty: true
        resources:
        limits:
            # pty-device-plugin won't allocate more than one pty to any pod
            # always set to 1
            arhat.dev/pty: 1
    restartPolicy: Never
   ```

3. Attach to host pty pod with `kubectl attach`, (for above example pod, `kubectl attach -it pty-client-at-my-kube-node`)

## TODO

- Build a `Kubernetes` operator to restrict Linux system user in resource request

__NOTICE__: This is one of my hobby projects, due to lack of hours in a day, items in this TODO list can be slow to happen

## LICENSE

[![GitHub license](https://img.shields.io/github/license/arhat-dev/kube-host-pty.svg)](https://github.com/arhat-dev/kube-host-pty/blob/master/LICENSE.txt)

```text
Copyright arhat.dev (github.com/arhat-dev)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
