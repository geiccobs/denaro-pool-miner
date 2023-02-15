package main

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net"
	"pool-miner/sha256"
	"strings"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

var (
	processes = new(sync.Map) // here we'll store routines statuses

	// everything here is pretty standard, if something doesn't work you should check your network first
	client = &fasthttp.Client{
		MaxConnDuration: time.Second * 30,
		ReadTimeout:     time.Second * 30,
		WriteTimeout:    time.Second * 30,
		Dial: func(addr string) (net.Conn, error) {
			return fasthttp.DialTimeout(addr, time.Second*5)
		},
	}

	ADDRESS          = "" // E1o5MVMtHLMys1fKutkSYWFqoh2o5iGLtXB487fifJa9V donations are always accepted
	WORKERS          = 4  // concurrent workers to spawn
	SHARE_DIFFICULTY = 6  // share difficulty (pretty self descriptive)

	NODE_URL = "https://denaro-node.gaetano.eu.org/" // down 24/7
	POOL_URL = "https://denaro-pool.gaetano.eu.org/" // never down, nobody knows it
)

func getTransactionsMerkleTree(transactions []string) string {

	var fullData []byte

	for _, transaction := range transactions {
		data, _ := hex.DecodeString(transaction)
		fullData = append(fullData, data...)
	}

	hash := sha256.New()
	if _, err := hash.Write(fullData); err != nil {
		panic(err)
	}

	return hex.EncodeToString(hash.Sum(nil))
}

func checkBlockIsValid(blockHash string, shareChunk string, chunk string, idifficulty int, charset string, hasDecimal bool) (bool, bool) {
	if strings.HasPrefix(blockHash, shareChunk) {
		if hasDecimal {
			return true, strings.HasPrefix(blockHash, chunk) && strings.Contains(charset, string(blockHash[idifficulty]))
		} else {
			return true, strings.HasPrefix(blockHash, chunk)
		}
	}
	return false, false
}

func worker(start int, step int, res MiningInfoResult, address string) {

	var difficulty = res.Difficulty
	var idifficulty = int(difficulty)
	var shareDifficulty = SHARE_DIFFICULTY

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from: %v\n", r)
			stopWorkers()
			return
		}
	}()

	_, decimal := math.Modf(difficulty)

	lastBlock := res.LastBlock
	if lastBlock.Hash == "" {
		var num uint32 = 30_06_2005

		data := make([]byte, 32)
		binary.LittleEndian.PutUint32(data, num)

		lastBlock.Hash = hex.EncodeToString(data)
	}

	chunk := lastBlock.Hash[len(lastBlock.Hash)-idifficulty:]

	var shareChunk string

	if shareDifficulty > idifficulty {
		shareDifficulty = idifficulty
	}
	shareChunk = chunk[:shareDifficulty]

	charset := "0123456789abcdef"
	if decimal > 0 {
		count := math.Ceil(16 * (1 - decimal))
		charset = charset[:int(count)]
	}

	addressBytes := stringToBytes(address)
	t := float64(time.Now().UnixMicro()) / 1000000.0
	i := start
	a := time.Now().Unix()
	txs := res.PendingTransactionsHashes
	merkleTree := getTransactionsMerkleTree(txs)

	if start == 0 {
		log.Printf("Difficulty: %f\n", difficulty)
		log.Printf("Block number: %d\n", lastBlock.Id)
		log.Printf("Confirming %d transactions\n", len(txs))
	}

	var prefix []byte
	dataHash, _ := hex.DecodeString(lastBlock.Hash)
	prefix = append(prefix, dataHash...)
	prefix = append(prefix, addressBytes...)
	dataMerkleTree, _ := hex.DecodeString(merkleTree)
	prefix = append(prefix, dataMerkleTree...)
	dataA := make([]byte, 4)
	binary.LittleEndian.PutUint32(dataA, uint32(a))
	prefix = append(prefix, dataA...)
	dataDifficulty := make([]byte, 2)
	binary.LittleEndian.PutUint16(dataDifficulty, uint16(difficulty*10))
	prefix = append(prefix, dataDifficulty...)

	if len(addressBytes) == 33 {
		data1 := make([]byte, 2, 2)
		binary.LittleEndian.PutUint16(data1, uint16(2))

		oldPrefix := prefix
		prefix = data1[:1]
		prefix = append(prefix, oldPrefix...)
	}

	var _hex = prefix
	var prefixHash = sha256.New()
	if _, err := prefixHash.Write(prefix); err != nil {
		panic(err)
	}

	for {
		found := true
		check := 5000000 * step

	checkLoop:
		for {
			if process, ok := processes.Load(start); !ok || !process.(Goroutine).Alive {
				return
			}

			dataI := make([]byte, 4)
			binary.LittleEndian.PutUint32(dataI, uint32(i))

			fullHash := prefixHash
			if _, err := fullHash.Write(dataI); err != nil {
				panic(err)
			}
			shareValid, blockValid := checkBlockIsValid(hex.EncodeToString(fullHash.Sum(nil)), shareChunk, chunk, idifficulty, charset, decimal > 0)

			if shareValid {
				var reqP Share

				_hex = _hex[:len(prefix)]
				_hex = append(_hex, dataI...)

				req := POST(
					POOL_URL+"share",
					map[string]interface{}{
						"block_content":    hex.EncodeToString(_hex),
						"txs":              txs,
						"id":               lastBlock.Id + 1,
						"share_difficulty": difficulty,
					},
				)
				_ = json.Unmarshal(req.Body(), &reqP)

				if reqP.Ok {
					log.Println("SHARE ACCEPTED")
				} else {
					log.Println("SHARE NOT ACCEPTED")
					log.Println(string(req.Body()))
					stopWorkers()
					return
				}
			}

			if blockValid {
				break checkLoop
			}

			i = i + step
			if (i-start)%check == 0 {
				elapsedTime := float64(time.Now().UnixMicro())/1000000.0 - t
				log.Printf("Worker %d: %dk hash/s", start+1, i/step/int(elapsedTime)/1000)

				if elapsedTime > 90 {
					found = false
					break checkLoop
				}
			}
		}

		if found {
			var reqP PushBlock

			log.Println(hex.EncodeToString(_hex))

			req := POST(
				NODE_URL+"push_block",
				map[string]interface{}{
					"block_content": hex.EncodeToString(_hex),
					"txs":           txs,
					"id":            lastBlock.Id + 1,
				},
			)
			_ = json.Unmarshal(req.Body(), &reqP)

			if reqP.Ok {
				log.Println("BLOCK MINED")
			}

			stopWorkers()
			return
		}
	}
}

func main() {

	flag.StringVar(&ADDRESS, "address", ADDRESS, "address that'll receive mining rewards")
	flag.IntVar(&WORKERS, "workers", WORKERS, "number of concurrent workers to spawn")
	flag.StringVar(&NODE_URL, "node", NODE_URL, "node to which we'll retrieve mining info")
	flag.StringVar(&POOL_URL, "pool", POOL_URL, "pool to which we'll mine on")
	flag.IntVar(&SHARE_DIFFICULTY, "share_difficulty", SHARE_DIFFICULTY, "pretty self descriptive")

	flag.Parse()

	// ask for address if not inserted as flag
	if len(ADDRESS) == 0 {
		fmt.Print("Insert your address (available at https://t.me/DenaroCoinBot): ")
		if _, err := fmt.Scan(&ADDRESS); err != nil {
			panic(err)
		}
	}

	var reqP MiningAddress

	req := GET(
		POOL_URL+"get_mining_address",
		map[string]interface{}{
			"address": ADDRESS,
		},
	)

	if err := json.Unmarshal(req.Body(), &reqP); err != nil {
		panic(err)
	}

	miningAddress := reqP.Address
	log.Println(miningAddress)

	for {
		log.Printf("Starting %d workers", WORKERS)

		var reqP MiningInfo

		req := GET(NODE_URL+"get_mining_info", map[string]interface{}{})
		_ = json.Unmarshal(req.Body(), &reqP)

		for _, i := range createRange(1, WORKERS) {
			log.Printf("Starting worker n.%d", i)
			go worker(i-1, WORKERS, reqP.Result, miningAddress)

			processes.Store(i-1, Goroutine{Id: i - 1, Alive: true})
		}

		elapsedSeconds := 0

	waitLoop:
		for allAliveWorkers() {
			time.Sleep(1 * time.Second)
			elapsedSeconds += 1

			if elapsedSeconds > 180 {
				stopWorkers()
				break waitLoop
			}
		}
	}
}
