package main

import (
	"errors"
	"fmt"
	"github.com/daoleno/uniswap-sdk-core/entities"
	"github.com/daoleno/uniswapv3-sdk/constants"
	"github.com/daoleno/uniswapv3-sdk/examples/contract"
	"github.com/daoleno/uniswapv3-sdk/examples/helper"
	"github.com/daoleno/uniswapv3-sdk/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
)

var (
	// Asset1 , Asset2 - Setting pool assets
	Asset1 = WMATIC
	Asset2 = WETH
	// Fee Setting pool fee: FeeLowest - 0.01%, FeeLow - 0.05%, FeeMedium - 0.3%, FeeHigh - 1%
	Fee = constants.FeeMedium
)

func main() {
	// connecting to blockchain rpc server
	client, err := ethclient.Dial(helper.PolygonRPC)
	if err != nil {
		panic(err)
	}

	// calculating pool address
	poolAddress, err := GetPoolAddress(client, Asset1.Address, Asset2.Address, new(big.Int).SetUint64(uint64(Fee)))
	if err != nil {
		panic(err)
	}
	fmt.Println("PoolAddress:", poolAddress)

	// creating instance to make calls
	contractPool, err := contract.NewUniswapv3Pool(poolAddress, client)
	if err != nil {
		panic(err)
	}

	// retrieving tick Spacing used for that pool
	tickSpacing, err := contractPool.TickSpacing(nil)
	if err != nil {
		panic(err)
	}
	TickSpacing := tickSpacing.Int64()
	fmt.Printf("TickSpacing: %+v\n", TickSpacing)

	// retrieving Liquidity of the current tick range
	liquidity, err := contractPool.Liquidity(nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Liquidity: %v\n", liquidity)

	// retrieving current Tick
	slot0, err := contractPool.Slot0(nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("slot0: current Tick: %v\n", slot0.Tick)

	// looking for boundaries of current ticks range
	lowerTick, upperTick := FindBoundaries(slot0.Tick.Int64(), TickSpacing)
	fmt.Printf("lowerTick: %v, upperTick: %v\n", lowerTick, upperTick)

	// Calculate root of the price √pa
	sqrtRatioAX96, err := utils.GetSqrtRatioAtTick(int(upperTick))
	if err != nil {
		log.Fatal(err)
	}
	priceA := GetSqr192(sqrtRatioAX96, Asset1, Asset2)
	fmt.Printf("sqrtRatioAX96: %v (real price: %s)\n", sqrtRatioAX96.Text(10), priceA)
	RatioA := new(big.Int).Rsh(sqrtRatioAX96, 64)

	// Calculate root of the price √P
	sqrtRatioCurrentX96, err := utils.GetSqrtRatioAtTick(int(slot0.Tick.Int64()))
	if err != nil {
		log.Fatal(err)
	}
	priceCurr := GetSqr192(sqrtRatioCurrentX96, Asset1, Asset2)
	fmt.Printf("sqrtRatioCurreX96: %v (real price: %s)\n", sqrtRatioCurrentX96, priceCurr)
	RatioCurrent := new(big.Int).Rsh(sqrtRatioCurrentX96, 64)

	// Calculate root of the price √pb
	sqrtRatioBX96, err := utils.GetSqrtRatioAtTick(int(lowerTick))
	if err != nil {
		log.Fatal(err)
	}
	priceB := GetSqr192(sqrtRatioBX96, Asset1, Asset2)
	fmt.Printf("sqrtRatioBX96: %v (real price: %s)\n", sqrtRatioBX96, priceB)
	RatioB := new(big.Int).Rsh(sqrtRatioBX96, 64)

	// now use Equations from chapter 3.3.3 of document https://atiselsts.github.io/pdfs/uniswap-v3-liquidity-math.pdf
	// to calculate the amount of the assets in a ticks range
	temp := new(big.Int).Div(new(big.Int).Sub(RatioB, RatioCurrent), new(big.Int).Mul(RatioCurrent, RatioB))
	Asset1Amount := new(big.Int).Mul(liquidity, temp)

	Asset2Amount := new(big.Int).Mul(liquidity, new(big.Int).Sub(RatioCurrent, RatioA))

	fmt.Printf("\nAvailable assets in ticks range: %s: %v, %s: %v\n", Asset1.Symbol(), Asset1Amount, Asset2.Symbol(), Asset2Amount)
}

// GetSqr192 print a human-readable value of price Asset1/Asset2 for passed Tick
func GetSqr192(sqrtRatioX96 *big.Int, asset1, asset2 entities.Currency) string {
	ratioX192 := new(big.Int).Mul(sqrtRatioX96, sqrtRatioX96)
	price := entities.NewPrice(asset1, asset2, ratioX192, constants.Q192)
	return price.ToFixed(10)
}

// GetPoolAddress calculates pool address
func GetPoolAddress(client *ethclient.Client, token0, token1 common.Address, fee *big.Int) (common.Address, error) {
	f, err := contract.NewUniswapv3Factory(common.HexToAddress(ContractV3Factory), client)
	if err != nil {
		return common.Address{}, err
	}
	poolAddr, err := f.GetPool(nil, token0, token1, fee)
	if err != nil {
		return common.Address{}, err
	}
	if poolAddr == (common.Address{}) {
		return common.Address{}, errors.New("pool is not exist")
	}

	return poolAddr, nil
}

// FindBoundaries looks for lower and upper tick number of range contained current tick number
func FindBoundaries(currentTick int64, tickSpacing int64) (int64, int64) {
	remainder := currentTick % tickSpacing
	if remainder < 0 {
		remainder += tickSpacing
	}

	lowerTick := currentTick - remainder
	upperTick := lowerTick + tickSpacing
	return lowerTick, upperTick
}
