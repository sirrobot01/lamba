## Lamba(A self-hosted [AWS Lambda](https://aws.amazon.com/lambda/) clone)

#### WIP

Lamba is a self-hosted AWS Lambda clone written in Go. It is designed to be compatible with the AWS Lambda API, but it is not a drop-in replacement

This only supports the Go runtime at the moment.

### Features

- [x] Adding a function
- [x] Invoking a function via
    - [x] HTTP
    - [x] Direct CLI invocation
- [x] List functions
- [x] List runtimes


### Installation

#### From Binary

You can download the pre-built binaries for different platforms from the [Releases](https://github.com/sirrobot01/lamba/releases/) page. Extract them using tar, move it to your $PATH and you are ready to go.

```bash

#### From Source

```bash
go get -u github.com/sirrobot01/lamba
```

### Usage

```bash
lamba --help
```

### Examples

You can see a sample function in the `examples` directory

#### Adding a function from a go file

```bash
lamba add --name hello --runtime go --handler Handler --file hello.go
```

#### Adding a function from a directory

```bash
lamba add --name hello --runtime go --handler Handler --file .
```

#### Invoking a function

```bash
lamba invoke --function hello --payload '{"name": "world"}'
```

Check [CLI.md](docs/cli.md) for more information

### Roadmap

- [ ] Add support for more runtimes(Python, Node.js, etc.)
- [ ] Add support for more event sources
- [ ] Add support for more triggers