# GotIt Tunnel

[![Go Report Card](https://goreportcard.com/badge/github.com/AlexEkdahl/gotit)](https://goreportcard.com/report/github.com/AlexEkdahl/gotit)
[![GitHub license](https://img.shields.io/github/license/AlexEkdahl/gotit)](https://github.com/AlexEkdahl/gotit/blob/main/LICENSE)


## Installation

Before you start, make sure you have a working Go environment. See the [install instructions](http://golang.org/doc/install.html).

To install GotIt, simply run:

```shell
go get github.com/AlexEkdahl/gotit
```

## Usage

Set the port numbers for the HTTP and SSH servers using the `HTTP_PORT` and `SSH_PORT` environment variables respectively. If not set, the HTTP server defaults to port 3000 and the SSH server to port 2222.

```shell
export HTTP_PORT=<<HTTP_PORT>>
export SSH_PORT=<<SSH_PORT>>
```

Then, start the servers:

```shell
go run main.go
```

## Contributing

We welcome contributions from the community. Please read the [contributing guide](CONTRIBUTING.md) before getting started.

## License

This project is licensed under the terms of the MIT license. See the [LICENSE](LICENSE) file for details.

