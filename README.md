# pyp - npm for Python

## What is pyp?

`pyp` is a command-line interface (CLI) tool that enhances Python project and dependency management. It is a wrapper around [**`pip`**](https://github.com/pypa/pip), providing functionality similar to what **npm** offers for Node.js. Inspired by npm's ease of use, `pyp` aims to streamline Python development workflows.

**NOTE:** While `pyp` offers an npm-like experience for Python, it's not the recommended tool for managing Python projects. More robust alternatives exist, such as [Pipenv](https://pipenv.pypa.io/) and [Poetry](https://python-poetry.org/). `pyp` was developed as a hobby project primarily for educational purposes.

## Requirements

To use `pyp` you need the following:

- [Python](https://www.python.org/)
- [pip](https://pypi.org/project/pip/)
- [venv](https://docs.python.org/3/library/venv.html)

## Build and Installation

To build and install `pyp`, you have two options:

1. Local Build:
   If you have Go installed on your system, you can build `pyp` locally:

   - Run `go build` in the repository directory.
   - This will create a `pyp` executable in the same directory.

2. System-wide Installation:
   To install `pyp` system-wide:
   - Run `go install` in the repository directory.
   - This will compile and install `pyp` to your system's Go bin directory.
   - Ensure your Go bin directory is in your system's PATH.

For more information about `go install`, refer to the [official Go documentation](https://go.dev/ref/mod#go-install).
