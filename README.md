# BetterBetter Sports Betting Prediction Pipeline

This repository provides a pipeline for fetching, modeling, optimizing bets, and backtesting sports betting strategies using team and player statistics. It is tailored for basketball data, specifically the NBA, and uses models such as Bayesian Autoregressive and Energy models.

## Prerequisites

Ensure that you have the required dependencies installed:

1. **Go**: Install Go programming language and set up the Go environment.
2. **betterbetter**: A custom CLI tool used for fetching data, generating predictions, optimizing bets, and backtesting.

### Setup

To set up the environment:

```bash
export PATH=$PATH:~/go/bin
go install
```

## Fetching Data
### Fetch Player and Team Statistics
#### Fetch statistics for the specified teams and season year:

```
betterbetter fetchdata -s nba -t celtics,lakers -y 2023
```

### Fetch Odds for Each Game
#### Fetch odds for each game of the specified teams and season:

```
betterbetter fetchodds -s nba -t celtics,lakers -d date
```

## Modeling

```
betterbetter makepredictions 
```

-b flag for bayesian modeling
-e flag for energy based modeling

## Betslip Optimization
### MaxiMax problem 
Want to maximize profit & arbitrage opportunity

```
betterbetter optimizebets
```

## Backtesting

```
betterbetter backtest
```
