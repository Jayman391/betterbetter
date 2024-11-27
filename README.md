### flow of project

scrape player and team data
scrape player and team odds

'''
betterbetter fetchdata -s -t -p -y
betterbetter fetchodds -s -d
'''

build regression model and forecast probability distributions of metrics for each team and player
compare predicted probabilities to odds probabilities

'''
betterbetter model
'''

calculate differentials between predicted and actual
average differentials across sportsbooks

'''
betterbetter arbitrage
'''

set risk reward ratio
create parlays via combinations of bets
  calculate differentials on parlays
    universe includes individual bets and parlays all with expected values
use differentials and expected values for each bet in universe
run optimization routine to maximize expected value given risk/reward constraint
  make (multiple) sets of bets that satisfy the constraints

'''
betterbetter makebets -r
'''


backtest each set of bets and calculate expected profit
calculate average % profit for risk reward scheme
make betslips for new games
'''
betterbetter predict
'''

# in run_pipeline.sh

'''
betterbetter fetchdata -s "" -t ",,," -p ",,," -y "YYYY"
betterbetter fetchodds -s "" -d "YYYY-MM-DD"
betterbetter model
betterbetter arbitrage
betterbetter makebets -r ""
betterbetter predict
'''

