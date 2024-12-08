package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type GetPrequalParametersResponse struct {
	Success bool
	Data    PrequalParametersResponse
}

func GetPrequalParameters() map[string]any {
	result := make(map[string]any)
	result["max_life_time"] = 5
	result["pool_size"] = 16
	result["probe_factor"] = 1.2
	result["probe_remove_factor"] = 1
	result["mu"] = 1

	res, err := http.Get(fmt.Sprintf("%s/%s", adminUrl, "get-prequal-parameters"))

	if err != nil {
		log.Print("Error getting prequal parameters. Using default values...")
		return result
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Print("Error getting prequal parameters. Using default values...")
		// log.Panicf("received non-200 response: %d", res.StatusCode)
		return result
	}

	body, err := io.ReadAll(res.Body)

	if err != nil {
		log.Print("Error getting prequal parameters. Using default values...")
		// log.Panicf("error reading response body %v", err)
	}

	var prequalParameterResponse GetPrequalParametersResponse

	err = json.Unmarshal([]byte(body), &prequalParameterResponse)
	if err != nil {
		// log.Print("Error getting prequal parameters. Using default values...")
		log.Panicf("error parsing JSON: %v", err)
	}

	// get the url from the response body and put it in an array
	result["id"] = prequalParameterResponse.Data.Id
	result["max_life_time"] = prequalParameterResponse.Data.MaxLifeTime
	result["pool_size"] = prequalParameterResponse.Data.PoolSize
	result["probe_factor"] = prequalParameterResponse.Data.ProbeFactor
	result["probe_remove_factor"] = prequalParameterResponse.Data.ProbeRemoveFactor
	result["mu"] = prequalParameterResponse.Data.Mu

	return result
}
