package main

func calculateChange(original, new, price float32) float32 {
	return price * (original - new)
}

func initTokensCheckpoint(height uint64) map[string]uint64 {
	result := make(map[string]uint64)

	result[USDC] = height
	result[USDT] = height
	result[BTC] = height
	result[ETH] = height
	result[BNB] = height
	result[LTC] = height
	result[ZEC] = height
	result[MATIC] = height
	result[DASH] = height
	result[DAI] = height
	result[XMR] = height

	return result
}
