# doh-relay &middot; [![License](https://img.shields.io/hexpm/l/plug?logo=Github&style=flat)](LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/tinkernels/sharp)](https://goreportcard.com/report/github.com/tinkernels/sharp)
sharp is a reverse proxy with tls terminator

- Multiple source addresses to multiple upstreams

- TLS terminating

## Build

```
make release
```

## Usage 

```
sharp -config config.yml
```
### Config file
  There's a example config file with comments [here](config-example.yml).

## License

[Apache-2.0](https://github.com/tinkernels/doh-relay/blob/master/LICENSE)
