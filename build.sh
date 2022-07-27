DIR="builds/"

# remove builds dir if exists
if [ -d "$DIR" ]; then
    rm -r $DIR
fi

# compile for linux
# 64 bit
GOOS=linux GOARCH=amd64 go build -o $DIR/pool-miner-linux64
# 32 bit
GOOS=linux GOARCH=386 go build -o $DIR/pool-miner-linux32

echo "compiled for linux"

# compile for windows
# 64 bit
GOOS=windows GOARCH=amd64 go build -o $DIR/pool-miner-win64.exe
# 32 bit
GOOS=windows GOARCH=386 go build -o $DIR/pool-miner-win32.exe

echo "compiled for windows"

# compile for macos
# 64 bit
GOOS=darwin GOARCH=amd64 go build -o $DIR/pool-miner-macos64

echo "compiled for macos"