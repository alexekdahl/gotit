# gotit
[![Go Report Card](https://goreportcard.com/badge/github.com/AlexEkdahl/gotit)](https://goreportcard.com/report/github.com/AlexEkdahl/gotit)
[![GitHub license](https://img.shields.io/github/license/AlexEkdahl/gotit)](https://github.com/AlexEkdahl/gotit/blob/main/LICENSE)

**gotit** is a simple and efficient tool that establishes a secure peer-to-peer connection between an SSH client and an HTTP server. It's designed to be lightweight and easy to use, making it perfect for quick data transfers or remote command execution.

## Features

- Secure SSH connection
- Lightweight HTTP server
- Peer-to-peer data transfer
- No data storage on the server

## Getting Started

To get started with **gotit**, you'll need to clone the repository and build the project.

```bash
git clone https://github.com/AlexEkdahl/gotit.git
cd gotit
make
```

## Usage

You can start the **gotit** server with the following command:

```bash
make run
```

By default, the HTTP server listens on port 8080 and the SSH server listens on port 2222. You can change these settings using the httpport and sshport flags.

```bash
gotit --httpport 8000 --sshport 2200
```

## Sending Data

To send data, you'll need to establish an SSH connection to the Gotit server. Once connected, the server will provide a unique URL for the session.

You can then initiate a file transfer from the SSH client. The data you send will be forwarded to the HTTP client that connects to the provided URL.

For example, you can use the following commands to send a file:
Once the server is running, you can establish an SSH connection to it. The server will provide a unique URL for each SSH session. You can then use this URL to send HTTP requests to the SSH client.

```bash
ssh sshserver -p 2020 < main.go
```
or
```bash
cat main.go | ssh sshserver -p 2020
```

## Contributing

Contributions to **gotit** are welcome! Please submit a pull request or create an issue if you have any improvements or bug fixes.

## License

This project is licensed under the terms of the MIT license. See the [LICENSE](LICENSE) file for details.
