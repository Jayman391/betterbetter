## Flow of Project

1. Scrape player and team data:
  - `-s`: sport
  - `-t`: teams
  - `-p`: players
  - `-y`: year

  Example command: `betterbetter fetchdata -s -t -p -y`

2. Scrape player and team odds:
  - `-s`: sport
  - `-d`: date

  Example command: `betterbetter fetchodds -s -d`

3. Build regression model and forecast probability distributions of metrics for each team and player. Compare predicted probabilities to odds probabilities:
  - `-l`: lags for AR model
  - `-c`: chains for Bayesian sampler
  - `-e`: examples for posterior predictive
  - `-s`: number of samples from posterior predictive

  Example command: `betterbetter bayes -l -c -e -s`

4. Calculate differentials between predicted and actual. Average differentials across sportsbooks:
  - `-s`: path to posterior predictions
  - `-o`: path to odds data

  Example command: `betterbetter arbitrage -s -o`

5. Set risk reward ratio. Create parlays via combinations of bets. Calculate differentials on parlays. The universe includes individual bets and parlays, all with expected values. Use differentials and expected values for each bet in the universe. Run optimization routine to maximize expected value given risk/reward constraint. Make sets of bets that satisfy the constraints:
  - `-r`: risk reward ratio
  - `-m`: maximum number of bets to return

  Example command: `betterbetter makebets -r -m`

6. Backtest each set of bets and calculate expected profit. Calculate average % profit for risk reward scheme. Make betslips for new games:

  Example command: `betterbetter predict`

7. Look at `run_pipeline.sh` to set up the full pipeline with associated directories.
