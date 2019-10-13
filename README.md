# csc

CheckSum-based Content management

## Install

```sh
go get -u github.com/taskie/csc/cmd/csc
go get -u github.com/taskie/csc/cmd/cscman
```

## Usage

### csc

```sh
csc scan .
csc path "$PWD"
csc sha256 ff
csc find ./foo.txt
```

### cscman

```sh
cscman register bar example.local:csc.db
cscman sync bar
```

## License

Apache License 2.0
