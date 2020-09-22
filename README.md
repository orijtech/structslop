# structslop

[![Build Status](https://travis-ci.com/orijtech/structslop.svg?token=zRGT22WqV6Do9u8mxebC&branch=master)](https://travis-ci.com/orijtech/structslop)

Package structslop defines an [Analyzer](analyzer_link) that checks struct can be re-arranged fields to get optimal struct size.

## Installation

```sh
go get github.com/orijtech/structslop/cmd/structslop
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