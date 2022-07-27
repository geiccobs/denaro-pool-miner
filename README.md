# denaro cpu pool miner

## installation

```bash
git clone https://github.com/geiccobs/denaro-pool-miner
cd denaro-pool-miner
```

### compiling by source

You can skip this if you wanna use pre-built binary.  
[Install golang first](https://go.dev/doc/install)
```bash
go build .
```

## usage

```bash
./pool-miner -address youraddress -workers corescount -share_difficulty 6
```

`share_difficulty` should be adjusted according to your hashrate: if you see a lot of shares accepted, increment it.  
  
Use `./pool-miner -help` to see the full list of arguments
