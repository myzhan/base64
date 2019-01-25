# base64 [![Build Status](https://travis-ci.org/myzhan/base64.svg?branch=master)](https://travis-ci.org/myzhan/base64) [![Go Report Card](https://goreportcard.com/badge/github.com/myzhan/base64)](https://goreportcard.com/report/github.com/myzhan/base64) [![Coverage Status](https://codecov.io/gh/myzhan/base64/branch/master/graph/badge.svg)](https://codecov.io/gh/myzhan/base64)

## Description

This is a golang wrapper for [aklomp's awesome base64 library](https://github.com/aklomp/base64).

I wrote this to learn cgo. It's not well-tested and production ready.

## Build

The Makefile of the C library is modified slightly to generate a libbase64.a file, which will be linked by cgo.

```bash
cd deps/base64 && make
```

## Known Issues

- Whitespace is not skipped, decoding strings like "c3VyZQ==\r" will fail.