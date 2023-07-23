# gotit
[![Go Report Card](https://goreportcard.com/badge/github.com/AlexEkdahl/gotit)](https://goreportcard.com/report/github.com/AlexEkdahl/gotit)
[![GitHub license](https://img.shields.io/github/license/AlexEkdahl/gotit)](https://github.com/AlexEkdahl/gotit/blob/main/LICENSE)
[![Build and Test](https://github.com/AlexEkdahl/gotit/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/AlexEkdahl/gotit/actions/workflows/test.yml)

<p align="left">
  <img src="https://img.gotit.sh/gotit-logo.png"  width="45%" />
</p>

**gotit** is a lightweight tool that establishes a secure peer-to-peer connection between an SSH client and an HTTP client. It's designed for quick data transfers without storing any data on the server.

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

To send data, establish an SSH connection to the **gotit** server. Once connected, the server will provide a unique URL for the session.
Initiate a file transfer from the SSH client. The data you send will be forwarded to the HTTP client that connects to the provided URL.

```bash
ssh sshserver < main.go
```
```bash
cat main.go | ssh sshserver
```

If you know the file type of the data you're transferring, you can include it as an argument in the command. For example, if you're transferring a JSON file, you can use the following command:

```bash
ssh sshserver json < data.json
```
```bash
cat main.go | ssh sshserver go
```

If you don't specify a file type type in the command, the **gotit** server will determine the MIME type from the first 512 bytes of the data. If the MIME type cannot be determined from the first 512 bytes, the server will default to application/octet-stream, which represents arbitrary binary data.

## Contributing

Contributions to **gotit** are welcome! Please submit a pull request or create an issue if you have any improvements or bug fixes.

## License

This project is licensed under the terms of the MIT license. See the [LICENSE](LICENSE) file for details.
