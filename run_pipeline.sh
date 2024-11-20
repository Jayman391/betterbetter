export PATH=$PATH:~/go/bin
go install

conda env create -f environment.yml
conda activate myenv

## Fetching Sports Data

betterbetter fetch player -p tatum,curry -s nba -y 2023 

## Modeling

betterbetter bayes --predict 

## Fetching Odds Data

betterbetter fetch odds -s nba 

## Create Report and Risk Analysis

betterbetter profile 

