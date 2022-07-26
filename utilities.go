package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/btcsuite/btcutil/base58"
	"github.com/valyala/fasthttp"
)

func stringToBytes(text string) []byte {

	var data []byte

	data, err := hex.DecodeString(text)
	if err != nil {
		data = base58.Decode(text)
	}

	return data
}

func createRange(min, max int) []int {

	var list = make([]int, max-min+1)

	for i := range list {
		list[i] = min + i
	}
	return list
}

func POST(url string, values map[string]interface{}) *fasthttp.Response {

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	req.SetRequestURI(url)
	req.SetConnectionClose()

	req.Header.SetMethod("POST")
	req.Header.SetContentType("application/json")

	kb, _ := json.Marshal(values)
	req.SetBody(kb)

	if err := client.Do(req, resp); err != nil {
		panic(err)
	}

	defer fasthttp.ReleaseRequest(req)

	return resp
}

func GET(url string, values map[string]interface{}) *fasthttp.Response {

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	url += "?"

	for key, value := range values {
		url += fmt.Sprintf("%s=%s&", key, value)
	}

	url = url[:len(url)-1]

	req.SetRequestURI(url)
	req.SetConnectionClose()

	req.Header.SetMethod("GET")
	req.Header.SetContentType("application/json")

	if err := client.Do(req, resp); err != nil {
		panic(err)
	}

	defer fasthttp.ReleaseRequest(req)

	return resp
}
