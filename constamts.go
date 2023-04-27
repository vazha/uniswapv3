package main

import (
	coreEntities "github.com/daoleno/uniswap-sdk-core/entities"
	"github.com/ethereum/go-ethereum/common"
)

const (
	PolygonRPC = "https://polygon-rpc.com/"

	PolygonChainID = 137
	MaticAddr      = "0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270"
	WETHAddr       = "0x7ceb23fd6bc0add59e62ac25578270cff1b9f619"
	UsdcAddr       = "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174"
	UsdtAddr       = "0xc2132d05d31c914a87c6611c10748aeb04b58e8f"
	AmpAddr        = "0x0621d647cecbFb64b79E44302c1933cB4f27054d"
	WBTCAddr       = "0x1bfd67037b42cf73acf2047067bd4f2c47d9bfd6"
)

const (
	ContractV3Factory            = "0x1F98431c8aD98523631AE4a59f267346ea31F984"
	ContractV3SwapRouterV1       = "0xE592427A0AEce92De3Edee1F18E0157C05861564"
	ContractV3SwapRouterV2       = "0x68b3465833fb72A70ecDF485E0e4C7bD8665Fc45"
	ContractV3NFTPositionManager = "0xC36442b4a4522E871399CD717aBDD847Ab11FE88"
	ContractV3Quoter             = "0xb27308f9F90D607463bb33eA1BeBb41C27CE5AB6"
)

var (
	MATIC = coreEntities.NewToken(PolygonChainID, common.HexToAddress(MaticAddr), 18, "MATIC", "Matic Network(PolyGon)")
	AMP   = coreEntities.NewToken(PolygonChainID, common.HexToAddress(AmpAddr), 18, "AMP", "Amp")
	USDC  = coreEntities.NewToken(PolygonChainID, common.HexToAddress(UsdcAddr), 6, "USDC", "USD Coin")
	USDT  = coreEntities.NewToken(PolygonChainID, common.HexToAddress(UsdtAddr), 6, "USDT", "Tether USD")
	WETH  = coreEntities.NewToken(PolygonChainID, common.HexToAddress(WETHAddr), 18, "WETH", "Wrapped ETH")
	WBTC  = coreEntities.NewToken(PolygonChainID, common.HexToAddress(WBTCAddr), 8, "WBTC", "Wrapped BTC")
)
