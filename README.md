## The solution of the task
In this example we use the Polygon blockchain.

To calculate the amount of assets we can buy without changing the price of the pair, we need to find the Liquidity <br>
that is in the range of lowerTick and upperTick based on the current Slot0.Tick and TickSpacing.

Knowing the current tick and also the upper and lower ticks, we should calculate the corresponding root of the price (√pa, √pb, √P)<br> to use further in the Equations.

Now use the Equations from chapter 3.3.3 of the document https://atiselsts.github.io/pdfs/uniswap-v3-liquidity-math.pdf <br>to calculate the amount for the each asset of the assets in a ticks range

We use library [github.com/daoleno/uniswapv3-sdk](github.com/daoleno/uniswapv3-sdk)   to make requests to smartcontract and make different Uniswap v3 mathematical operations

