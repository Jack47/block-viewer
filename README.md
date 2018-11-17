# block viewer

This is a block viewer, which reach webbtc.com to retrieve the binary of block (https://webbtc.com/block/<hash>.bin). It will:
1. Decode the header and display it to the terminal
2. Decode top 5 tx and display it to the terminal (shall include tx input / tx output)

## Usage
### Requirements
1. Go lang >= 1.8
2. jq (optional) to format json: yum install -y jq
### Build
If current dir is not named as `Jack47`, please rename it to `Jack47`, because that name is part of package name in `block.go`
```shell
mkdir -p $GOPATH/src/github.com/Jack47
cd $GOPATH/src/github.com/Jack47
git clone git@github.com:Jack47/block-viewer
cd ./block-viewer
go build -o block-viewer ./main.go
```
### View specific block
```shell
./block-viewer 000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f | jq '.'
```
## TODO
1. add hash, size field for each transaction
2. support other viewer using other format, such as hex
3. support paging when transactions are numerous
4. add more test blocks in test cases
