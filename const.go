package main

import "time"

const (
	DefaultBeaconCheckPoint uint64 = 1486744
	ChangeThreshold         int    = 1000
	PDEXV3Breakpoint        uint64 = 1000000000
)

const (
	beaconCheckInterval time.Duration = 40 * time.Second
)

const (
	USDT  string = "716fd1009e2a1669caacc36891e707bfdf02590f96ebd897548e8963c95ebac0"
	USDC  string = "1ff2da446abfebea3ba30385e2ca99b0f0bbeda5c6371f4c23c939672b429a42"
	DAI   string = "3f89c75324b46f13c7b036871060e641d996a24c09b3065835cb1d38b799d6c1"
	ETH   string = "ffd8d42dc40a8d166ea4848baf8b5f6e912ad79875f4373070b59392b1756c8f"
	BNB   string = "b2655152784e8639fa19521a7035f331eea1f1e911b2f3200a507ebb4554387b"
	BTC   string = "b832e5d3b1f01a4f0623f7fe91d6673461e1f5d37d91fe78c5c2e6183ff39696"
	XMR   string = "c01e7dc1d1aba995c19b257412340b057f8ad1482ccb6a9bb0adce61afbf05d4"
	LTC   string = "7450ad98cb8c967afb76503944ab30b4ce3560ed8f3acc3155f687641ae34135"
	DASH  string = "447b088f1c2a8e08bff622ef43a477e98af22b64ea34f99278f4b550d285fbff"
	ZEC   string = "a609150120c0247407e6d7725f2a9701dcbb7bab5337a70b9cef801f34bc2b5c"
	MATIC string = "dae027b21d8d57114da11209dce8eeb587d01adf59d4fc356a8be5eedc146859"
)

const (
	PDEXV2 = "pdexv2"
	PDEXV3 = "pdexv3"
)

const (
	prefixState                = "state"
	prefixCheckpointPoolAmount = "cpa"
	prefixPoolAmount           = "pa"
	prefixPriceHistory         = "pz"
	prefixChangeHistory        = "chg"
)

const (
	pdexV2RPC string = "getpdestate"
	pdexV3RPC string = "pdexv3_getState"
)

const (
	fullnodeURL     string = "https://mainnet.incognito.org/fullnode"
	binancePriceURL string = "https://api.binance.com/api/v3/ticker/price?symbol="
)

const (
	priceDAI   string = "DAIUSDT"
	priceETH   string = "ETHUSDT"
	priceBNB   string = "BNBUSDT"
	priceBTC   string = "BTCUSDT"
	priceXMR   string = "XMRUSDT"
	priceLTC   string = "LTCUSDT"
	priceDASH  string = "DASHUSDT"
	priceZEC   string = "ZECUSDT"
	priceMATIC string = "MATICUSDT"
)
