# structslop

[![Build Status](https://travis-ci.com/orijtech/structslop.svg?token=zRGT22WqV6Do9u8mxebC&branch=master)](https://travis-ci.com/orijtech/structslop)

Package structslop defines an [Analyzer](analyzer_link) that checks struct can be re-arranged fields to get optimal struct size.

## Installation

With Go modules:

```sh
go get github.com/orijtech/structslop/cmd/structslop
```

Without Go modules:

```sh
$ cd $GOPATH/src/github.com/orijtech/structslop
$ git checkout v0.0.1
$ go get
$ install ./cmd/structslop
```

## Usage

You can run `structslop` either on a Go package or Go files, the same way as
other Go tools work.

Example:

```sh
$ structslop github.com/orijtech/structslop/testdata/src/struct
```

or:

```sh
$ structslop ./testdata/src/struct/p.go
```

Sample output:

```text
/go/src/github.com/orijtech/structslop/testdata/struct/p.go:30:9: struct has size 24, could be 16, rearrange to struct{y uint64; x uint32; z uint32} for optimal size
/go/src/github.com/orijtech/structslop/testdata/struct/p.go:36:9: struct has size 40, could be 24, rearrange to struct{_ [0]func(); i1 int; i2 int; a3 [3]bool; b bool} for optimal size
/go/src/github.com/orijtech/structslop/testdata/struct/p.go:44:9: struct has size 32, could be 24, rearrange to struct{y uint64; z *httptest.Server; x uint32; t uint32} for optimal size
/go/src/github.com/orijtech/structslop/testdata/struct/p.go:51:9: struct has size 32, could be 24, rearrange to struct{y uint64; z *s; x uint32; t uint32} for optimal size
```

## Development

Go 1.15+

### Running test

Add test case to `testdata/src/struct` directory, then run:

```shell script
go test
```

## Contributing

TODO

[analyzer_link]: https://pkg.go.dev/golang.org/x/tools/go/analysis#Analyzer