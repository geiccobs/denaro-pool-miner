# Denaro CPU pool miner

## Installation

```bash
git clone https://github.com/geiccobs/denaro-pool-miner
cd denaro-pool-miner
```

### Compiling by source

You can skip this if you wanna use pre-built binary.  
[Install golang first](https://go.dev/doc/install)
```bash
go mod tidy
go build
```

## Usage

`share_difficulty` should be adjusted according to your hashrate: if you see a lot of shares accepted, increment it.  
  
Use `./pool-miner-{yourarchitecture} -help` to see the full list of arguments

### Running on Linux

```bash
cd builds/
./pool-miner-linux64 -address youraddress -workers corescount -share_difficulty 6
```

### Running on Windows

```bash
cd builds/
start pool-miner-win64.exe -address youraddress -workers corescount -share_difficulty 6
```

### Running on MacOS

```bash
cd builds/
./pool-miner-macos64 -address youraddress -workers corescount -share_difficulty 6
```
