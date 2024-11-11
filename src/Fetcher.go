package src

import (
	"fmt"
	"net/http"
  "io/ioutil"
)

func FetchData(sport string, requestType string, args map[string]string) {
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

  client := &http.Client {
  }
  req, err := http.NewRequest(method, url, nil)

  if err != nil {
    fmt.Println(err)
    return
  }
  req.Header.Add("x-rapidapi-key", "b6b0dbc354837ac6cfcaf07693d41da2")
  req.Header.Add("x-rapidapi-host", "v2.nba.api-sports.io")

  res, err := client.Do(req)
  if err != nil {
    fmt.Println(err)
    return
  }
  defer res.Body.Close()

  body, err := ioutil.ReadAll(res.Body)
  if err != nil {
    fmt.Println(err)
    return
  }
  fmt.Println(string(body))

}


