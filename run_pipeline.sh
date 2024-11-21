export PATH=$PATH:~/go/bin
go install

conda env create -f environment.yml
conda activate myenv

##### Fetching

## fetch player and team statistics

betterbetter fetchdata -s nba -t celtics,lakers -y 2023

## for each game fetch odds

betterbetter fetchodds -s nba -t celtics,lakers -d date

##### Modeling

## read team statistics and create timeseries data

    ## make variable lags for bayesian autoregressive model and energy model

betterbetter makepredictions 

    # -b flag for bayesian autoregressive model
    # -e flag for energy model

##### BetSlip Optimization

## MaxiMax optimization

    ## E[bet]
    ## Arbitrage Opportunity -> P(pred) - P(actual) 

betterbetter optimizebets 


##### Backtesting

## backtest the model

betterbetter backtest

    # -b flag for bayesian autoregressive model
    # -e flag for energy model






