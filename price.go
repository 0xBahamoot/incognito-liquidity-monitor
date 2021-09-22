package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

// func checkPrice() (map[string]float32, error) {
// 	httpClient := &http.Client{
// 		Timeout: time.Second * 10,
// 	}
// 	CG := coingecko.NewClient(httpClient)
// 	result := make(map[string]float32)
// 	ids := []string{priceBNB, priceBTC, priceDAI, priceDASH, priceETH, priceLTC, priceMATIC, priceUSDC, priceUSDT, priceXMR, priceZEC}
// 	datas, err := CG.SimplePrice(ids, []string{"usd"})
// 	if err != nil {
// 		return nil, err
// 	}
// 	for k, v := range *datas {
// 		switch k {
// 		case priceBNB:
// 			result[BNB] = (v["usd"])
// 		case priceBTC:
// 			result[BTC] = (v["usd"])
// 		case priceDAI:
// 			result[DAI] = (v["usd"])
// 		case priceDASH:
// 			result[DASH] = (v["usd"])
// 		case priceETH:
// 			result[ETH] = (v["usd"])
// 		case priceLTC:
// 			result[LTC] = (v["usd"])
// 		case priceMATIC:
// 			result[MATIC] = (v["usd"])
// 		case priceUSDC:
// 			result[USDC] = (v["usd"])
// 		case priceUSDT:
// 			result[USDT] = (v["usd"])
// 		case priceXMR:
// 			result[XMR] = (v["usd"])
// 		case priceZEC:
// 			result[ZEC] = (v["usd"])
// 		}
// 	}
// 	return result, nil
// }
func getPrice() (map[string]float32, error) {
	result := make(map[string]float32)
	ids := []string{priceBNB, priceBTC, priceDAI, priceDASH, priceETH, priceLTC, priceMATIC, priceXMR, priceZEC}
	for _, v := range ids {
		var price struct {
			Symbol string `json:"symbol"`
			Price  string `json:"price"`
		}
		resp, err := http.Get(binancePriceURL + v)
		if err != nil {
			return nil, err
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		err = json.Unmarshal(body, &price)
		if err != nil {
			log.Fatalln(err)
		}
		resp.Body.Close()
		value, err := strconv.ParseFloat(price.Price, 32)
		if err != nil {
			return nil, err
		}
		switch v {
		case priceBNB:
			result[BNB] = float32(value)
		case priceBTC:
			result[BTC] = float32(value)
		case priceDAI:
			result[DAI] = float32(value)
		case priceDASH:
			result[DASH] = float32(value)
		case priceETH:
			result[ETH] = float32(value)
		case priceLTC:
			result[LTC] = float32(value)
		case priceMATIC:
			result[MATIC] = float32(value)
		case priceXMR:
			result[XMR] = float32(value)
		case priceZEC:
			result[ZEC] = float32(value)
		}

	}
	result[USDC] = 1
	result[USDT] = 1

	return result, nil
}
