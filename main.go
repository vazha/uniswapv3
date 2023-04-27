package main

import (
	"errors"
	"fmt"
	"github.com/daoleno/uniswapv3-sdk/constants"
	"github.com/daoleno/uniswapv3-sdk/examples/contract"
	"github.com/daoleno/uniswapv3-sdk/examples/helper"
	"github.com/daoleno/uniswapv3-sdk/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math"
	"math/big"
)

var (
	// Asset1 , Asset2 - Setting pool assets
	//Asset1 = USDC
	//Asset2 = USDT

	//Asset1 = WBTC
	//Asset2 = WETH

	Asset1 = MATIC
	Asset2 = WETH

	// Fee Setting pool fee: FeeLowest - 0.01%, FeeLow - 0.05%, FeeMedium - 0.3%, FeeHigh - 1%
	Fee = constants.FeeLow
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
	priceA := PriceFromSqrtX96(sqrtRatioAX96, int(Asset1.Decimals()), int(Asset2.Decimals())).Text('f', 18)
	fmt.Printf("sqrtRatioAX96: %v (real price: %s)\n", sqrtRatioAX96.Text(10), priceA)

	// Calculate root of the price √P
	sqrtRatioCurrentX96, err := utils.GetSqrtRatioAtTick(int(slot0.Tick.Int64()))
	if err != nil {
		log.Fatal(err)
	}
	//sqrtRatioCurrentX96 = slot0.SqrtPriceX96 // todo (gives more precision?)
	priceC := PriceFromSqrtX96(sqrtRatioCurrentX96, int(Asset1.Decimals()), int(Asset2.Decimals())).Text('f', 18)
	fmt.Printf("sqrtRatioCX96: %v (real price: %s)\n", sqrtRatioCurrentX96, priceC)

	// Calculate root of the price √pb
	sqrtRatioBX96, err := utils.GetSqrtRatioAtTick(int(lowerTick))
	if err != nil {
		log.Fatal(err)
	}
	priceB := PriceFromSqrtX96(sqrtRatioBX96, int(Asset1.Decimals()), int(Asset2.Decimals())).Text('f', 18)
	fmt.Printf("sqrtRatioBX96: %v (real price: %s)\n", sqrtRatioBX96, priceB)

	if sqrtRatioAX96.Cmp(sqrtRatioBX96) >= 0 {
		sqrtRatioAX96, sqrtRatioBX96 = sqrtRatioBX96, sqrtRatioAX96
	}

	// now use Equations from chapter 3.3.3 of document https://atiselsts.github.io/pdfs/uniswap-v3-liquidity-math.pdf
	// to calculate the amount of the assets in a ticks range
	Liquidity := new(big.Float).SetInt(liquidity)

	Asset1Amount := new(big.Float).Mul(Liquidity, q64p96ToBigFloat(new(big.Int).Sub(sqrtRatioBX96, sqrtRatioCurrentX96)))
	Asset1Amount = new(big.Float).Quo(Asset1Amount, q64p96ToBigFloat(sqrtRatioCurrentX96))
	Asset1Amount = new(big.Float).Quo(Asset1Amount, q64p96ToBigFloat(sqrtRatioBX96))
	decimalAsset1 := new(big.Int).SetUint64(uint64(Asset1.Decimals()))
	decimal1 := new(big.Int).Exp(big.NewInt(10), decimalAsset1, nil)
	Asset1Amount = new(big.Float).Quo(Asset1Amount, new(big.Float).SetInt(decimal1)) // divide to floating point

	Asset2Amount := new(big.Float).Mul(Liquidity, q64p96ToBigFloat(new(big.Int).Sub(sqrtRatioCurrentX96, sqrtRatioAX96)))
	decimalAsset2 := new(big.Int).SetUint64(uint64(Asset2.Decimals()))
	decimal2 := new(big.Int).Exp(big.NewInt(10), decimalAsset2, nil)
	Asset2Amount = new(big.Float).Quo(Asset2Amount, new(big.Float).SetInt(decimal2)) // divide to floating point

	fmt.Printf("\nWe can buy %v %s or %v %s and it won't trigger an exit from the current ticks range.\n",
		Asset1Amount, Asset1.Symbol(), Asset2Amount, Asset2.Symbol())
	fmt.Printf("If we do not need to go beyond one tick, then we divide the number by another %v (tick range).",
		TickSpacing)
}

// PriceFromSqrtX96 print a human-readable value of price Asset1/Asset2 for passed sqrtPrice
func PriceFromSqrtX96(sqrtPriceX96 *big.Int, Asset1Decimals, Asset2Decimals int) *big.Float {
	sqrtPriceX96BigFloat := new(big.Float).SetInt(sqrtPriceX96)
	const q64_96ScalingFactor = float64(1 << 96)

	// Divide sqrtPriceX96 by the Q64.96 scaling factor
	sqrtPrice := new(big.Float).Quo(sqrtPriceX96BigFloat, new(big.Float).SetFloat64(q64_96ScalingFactor))

	// Square the obtained value to get the price
	price := new(big.Float).Mul(sqrtPrice, sqrtPrice)

	// Take into account the decimal places for the given token pair
	token0DecimalFactor := new(big.Float).SetFloat64(math.Pow10(Asset1Decimals))
	token1DecimalFactor := new(big.Float).SetFloat64(math.Pow10(Asset2Decimals))
	price = new(big.Float).Mul(price, new(big.Float).Quo(token0DecimalFactor, token1DecimalFactor))
	return price
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

// q64p96ToBigFloat convert Q64.96 format value to bog.float
func q64p96ToBigFloat(val *big.Int) *big.Float {
	floatValue := new(big.Float).SetInt(val)

	// Divide the floatValue by 2^96 to account for Q64.96 scaling
	pow96 := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(2), big.NewInt(96), nil))
	floatValue.Quo(floatValue, pow96)
	return floatValue
}
