# farely

[![Go](https://github.com/chermehdi/farely/actions/workflows/go.yml/badge.svg)](https://github.com/chermehdi/farely/actions/workflows/go.yml)

![Farely load balancer](./etc/farely.png)

A load balancer supporting multiple LB strategies written in Go.


## Goal

The goal of this project is purley educational, I started it as a brainstorming
session in [twitch.tv](https://twitch.tv/chermehdi), and I ended up having fun to the
point that I made it a project.


The balancer's main abstraction is a `Service`, each service has a name,
a balancing strategy, and a group of identical replicas that can serve requests
for the same service.

the service is configured via `yaml` file that is provided at startup, the most
recent example can be found in the `examples` folder.


```yaml
services: 
  - 
    matcher: /ui
    name: "Web UI"
    replicas: 
      - "http://192.168.23.1:8081"
      - "http://192.168.23.5:8082"
    strategy: RoundRobin
  - 
    matcher: /api/v1
    name: "Stateless API"
    replicas: 
      - "http://192.168.23.1:8081"
      - "http://192.168.23.5:8082"
    strategy: WeightedRoundRobin
```

## Building a demo

If you want to try load balancer, you can run the demo server in the
'cmd/demo/main.go' directory, which starts up a hello world server listening 
to a specifi port.


Launching the load balancer is as easy as 

```
go run cmd/farely/main.go --config-path /path/to/your/config/file
```

Or if you have an already built binary

```
./farely --config-path path/to/your/config
```


## Contributing

Creating issues, and Adding features are always welcome, I leave pointers in the
code for myself to remember to add tests / features, feel free to pick them up
and address them.

This is a personal project, that is provided for use without any support, and
for my own personal enjoyment, feel free to join me live to
chat in [twitch](https://twitch.tv/chermehdi) if you have questions or interested in contributing.
