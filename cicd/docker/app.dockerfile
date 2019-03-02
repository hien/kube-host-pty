# set `--build-arg TARGET` to the binary target name
FROM arhatdev/go-builder:onbuild as builder
FROM arhatdev/go-scratch:onbuild
