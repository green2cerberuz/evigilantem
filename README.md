# Evigilantem

Basic virtual chip8 emulator written in go.

## Pre requisites

- Vagrant

## Usage

1. Inside root folder run `vagrant up` to provision our virtual machine. 
2. After vm is ready run `vagrant ssh -- -X` to do x forwarding.

To run v8 chip code use:
`go run ./chip8.go`

To run a opengl window example, inside the share vagrant folder you can run:
`go run ./test.go`
If a black window open, development environment is setup correctly.
