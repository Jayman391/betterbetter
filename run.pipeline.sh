#!/bin/bash

export PATH=$PATH:~/go/bin
go install

year=2024
sport="nba"
teams="celtics,cavaliers,thunder,mavericks,warriors,grizzlies,nuggets,suns,magic,knicks,bucks,lakers"

# Fetch the initial data for the specified sport, teams, and year
betterbetter fetchdata -s "$sport" -t "$teams" -y "$year"

# Run Bayesian analysis
betterbetter bayes -l 2 -c 1 -e 2 -s 20000

# Initialize an empty list of dates
dates=()

# Split the teams by ',' and loop through each team
IFS=',' read -ra team_array <<< "$teams"
for team in "${team_array[@]}"; do
  # Read the data file for the team
  data_file="data/$sport/$year/$team/games.json"
  if [[ -f "$data_file" ]]; then
    # Extract the dates from the JSON file
    team_dates=$(jq -r '.response[]["date"]["start"]' "$data_file")
    # Append each date to the dates list
    for date in $team_dates; do
    # take first 10 characters of date
      dates+=("${date:0:10}")
    done
  else
    echo "Data file not found for team: $team"
  fi
done

# Remove duplicate dates
dates=($(printf "%s\n" "${dates[@]}" | sort -u))

# Process each date in the list
for d in "${dates[@]}"; do
  # if d does not start with 2024, skip
  if [[ ! "$d" =~ ^2024-11 ]]; then
    continue
  fi
  # Fetch odds for the specific date
  betterbetter fetchodds -s "$sport" -d "$d"

  # Loop through each team again for predictions and arbitrage
  for team in "${team_array[@]}"; do
    preds_dir="data/$sport/$year/$team/preds"
    output_dir="data/$sport/$year/$d"
    betterbetter arbitrage -s "$preds_dir" -o "$output_dir"
  done
done

# Make bets with the specified rules and limits
betterbetter makebets -r 10 -m 10000
