# vidmajster - A tiny webscraper for video sites

Vidmajster is a simple CLI webscraper for video sites.
I wanted to make a simple tool that can download videos from various sites
that don't have a native download option for some reason or use a custom
video player that doesn't allow you to download the video directly.

## Features

- Download multiple videos from various sites
- Simple CLI interface
- Lightweight and easy to use
- Supports multiple video formats (mp4, webm, ogg, mov, mkv, avi)

## Installation

You can download it from the [releases page](https://github.com/k-lar/vidmajster/releases).

Install vidmajster using go install:

```bash
go install github.com/k-lar/vidmajster@latest
```


## Usage

```bash
vidmajster --url <website url>
```

## Building from source

To build vidmajster from source, you need to have Go installed on your system.

1. Clone the repository:

```bash
git clone https://github.com/k-lar/vidmajster.git
cd vidmajster
```

2. Build the project:

```bash
go build -o vidmajster
```

You can install it with the Makefile like so:

```bash
# May require sudo
make install
```
