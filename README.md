# Lamba

A self-hosted alternative to AWS Lambda, written in Go.

![Home](docs/images/home.png)

## Features

- ✅ Function Management (Add, List, Invoke)
- ✅ Event Tracking
- ✅ Supports Python, Node.js, and Go
- ✅ Compatible with Containerd and Docker engines

## Installation

### Option 1: Binary
Download pre-built binaries from [Releases](https://github.com/sirrobot01/lamba/releases/)

### Option 2: From Source
```bash
go install github.com/sirrobot01/lamba@latest
```

## Quick Start

### Using Containerd
```bash
sudo ./lamba --engine containerd --port 8080
```

### Using Docker
```bash
./lamba --engine docker --port 8080
```

## Function Structure
```
function/
├── function.py
├── requirements.txt

# Create deployment package
zip -r function.zip function
```

## Prerequisites
- [Containerd](https://containerd.io/) (Note: Rootless installation requires manual service start)
  <br>OR
- [Docker](https://www.docker.com/)

## Roadmap
- Additional runtime support (Rust)
- Enhanced event sources
- Additional triggers

## Status
Project is under active development. Please report issues via GitHub.