package src

import (
	"fmt"
	"io"
	"net/http"
)

func FetchData(sport string, requestType string, args map[string]string) string {
	var url string = ""

	if sport == "nba"{
		fmt.Println("Fetching NBA data")
		url = "https://v2.nba.api-sports.io"
	}
	if sport == "nfl"{
		fmt.Println("Fetching NFL data")
		url = "https://v1.american-football.api-sports.io"
	}

	if requestType == "team"{
		url = url + "/teams?"
	}
	if requestType == "player"{
		url = url + "/players?"
	}
	if requestType == "team-stats" || requestType == "player-stats" {
		url = url + "/players/statistics?"
	}
	if requestType == "game"{
		url = url + "/games?"
	}


	fmt.Println(url)

	if len(args) == 0 {
		url = url[0:len(url)-1]
	} else {
		var index = 1
		for key, value := range args {
			if index == 1 {
				url = url + key + "=" + value
			} else {
				url = url + "&" + key + "=" + value
			}
			index = index + 1
		}
	}

	fmt.Println(url)

	
  method := "GET"

  client := &http.Client {}

  req, err := http.NewRequest(method, url, nil)

  if err != nil {
    fmt.Println(err)
  }
  req.Header.Add("x-rapidapi-key", "b6b0dbc354837ac6cfcaf07693d41da2")
  req.Header.Add("x-rapidapi-host", "v2.nba.api-sports.io")

  res, err := client.Do(req)
  if err != nil {
    fmt.Println(err) 
  }
  defer res.Body.Close()

  body, err := io.ReadAll(res.Body)
  if err != nil {
    fmt.Println(err)
  }
  
	return string(body)

}


func FetchGames(date string, sport string) string {
	url := ""
	
	if sport == "nba" {
		url = "https://api.the-odds-api.com/v4/historical/sports/basketball_nba/events?"
	}

	url += "apiKey=8a2e6be65caa1f5af89fca660c4e7eaa"

	url += "&date=" + date

	fmt.Println(url)

	client := &http.Client {}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		fmt.Println(err)
	}

	res, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(res)

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
  if err != nil {
    fmt.Println(err)
  }
  
	responseString := string(body)

	return responseString

}

func FetchOdds(date string, sport string, id string) string {
	url := ""

	if sport == "nba" {
		url = "https://api.the-odds-api.com/v4/historical/sports/basketball_nba/events"
	}

	url += "/" + id + "/odds?apiKey=8a2e6be65caa1f5af89fca660c4e7eaa"

	url += "&date=" + date + "&regions=us" + "&markets=player_points,player_rebounds,player_assists,player_blocks,player_steals,"

	url += "player_turnovers,h2h,spreads,totals,player_blocks_steals,player_points_rebounds,player_points_assists,"

	url += "player_rebounds_assists,player_points_rebounds_assists,player_first_basket,player_double_double,player_triple_double"

	fmt.Println(url)

	client := &http.Client {}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		fmt.Println(err)
	}

	res, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(res)

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
  if err != nil {
    fmt.Println(err)
  }
  
	responseString := string(body)

	return responseString	
}