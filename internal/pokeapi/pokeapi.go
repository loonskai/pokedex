package pokeapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	pokecache "github.com/loonskai/pokedexcli/internal/pokecache"
)

const API_BASEURL = "https://pokeapi.co/api/v2/"

type PokeapiClient struct {
	prev  *string
	next  *string
	cache pokecache.Cache[[]byte]
}

func (client *PokeapiClient) GetNext() (*APIPageResponseBody, error) {
	if client.next == nil {
		return nil, fmt.Errorf("No next page")
	}
	body, err := client.getBody(*client.prev)
	if err != nil {
		return nil, err
	}
	parsedBody, err := client.parsePageBody(body)
	if err != nil {
		return nil, err
	}
	updatePageLinks(client, *parsedBody)
	return parsedBody, nil
}

func (client *PokeapiClient) GetPrev() (*APIPageResponseBody, error) {
	if client.prev == nil {
		return nil, fmt.Errorf("No previous page")
	}
	body, err := client.getBody(*client.prev)
	if err != nil {
		return nil, err
	}
	parsedBody, err := client.parsePageBody(body)
	if err != nil {
		return nil, err
	}
	updatePageLinks(client, *parsedBody)
	return parsedBody, nil
}

func (client *PokeapiClient) GetFromLocationAreas(location string) (*APILocationAreasResponseBody, error) {
	parsedUrl, err := url.Parse(API_BASEURL + "location-area/" + location)
	if err != nil {
		return nil, err
	}
	body, err := client.getBody(parsedUrl.String())
	if err != nil {
		return nil, err
	}
	parsedBody, err := client.parseLocationBody(body)
	if err != nil {
		return nil, err
	}
	return parsedBody, nil
}

func (client *PokeapiClient) GetPokemonByName(name string) (*APIPokemonByNameResponseBody, error) {
	parsedUrl, err := url.Parse(API_BASEURL + "pokemon/" + name)
	if err != nil {
		return nil, err
	}
	body, err := client.getBody(parsedUrl.String())
	if err != nil {
		return nil, err
	}
	parsedBody, err := client.parsePokemonByNameBody(body)
	if err != nil {
		return nil, err
	}
	return parsedBody, nil
}

func (client *PokeapiClient) getBody(url string) ([]byte, error) {
	cached, ok := client.cache.Get(url)
	if ok {
		return cached, nil
	}
	body, err := get(url)
	if err != nil {
		return nil, err
	}
	client.cache.Add(url, body)
	return body, nil
}

func (client *PokeapiClient) parsePageBody(body []byte) (*APIPageResponseBody, error) {
	parsedBody := APIPageResponseBody{}
	err := json.Unmarshal(body, &parsedBody)
	if err != nil {
		return nil, err
	}
	return &parsedBody, nil
}

func (client *PokeapiClient) parseLocationBody(body []byte) (*APILocationAreasResponseBody, error) {
	parsedBody := APILocationAreasResponseBody{}
	err := json.Unmarshal(body, &parsedBody)
	if err != nil {
		return nil, err
	}
	return &parsedBody, nil
}

func (client *PokeapiClient) parsePokemonByNameBody(body []byte) (*APIPokemonByNameResponseBody, error) {
	parsedBody := APIPokemonByNameResponseBody{}
	err := json.Unmarshal(body, &parsedBody)
	if err != nil {
		return nil, err
	}
	return &parsedBody, nil
}

type APIPageResponseBody struct {
	Count   int              `json:"count"`
	Prev    string           `json:"previous"`
	Next    string           `json:"next"`
	Results []APIPokemonItem `json:"results"`
}

type APILocationAreasResponseBody struct {
	Id                   int                          `json:"id"`
	Name                 string                       `json:"name"`
	GameIndex            int                          `json:"game_index"`
	Location             APILocationItem              `json:"location"`
	Names                []APILocationNameItem        `json:"names"`
	EncounterMethodRates []APIEncounterMethodRateItem `json:"encounter_method_rates"`
	PokemonEncounters    []APIPokemonEncounterItem    `json:"pokemon_encounters"`
}

type APIPokemonItem struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type APILocationItem struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type APILocationNameItem struct {
	Name     string          `json:"name"`
	Language APILanguageItem `json:"language"`
}

type APILanguageItem struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type APIEncounterMethodRateItem struct {
	EncounterMethod APIEncounterMethodItem  `json:"encounter_method"`
	VersionDetails  []APIVersionDetailsItem `json:"version_details"`
}

type APIPokemonEncounterItem struct {
	Pokemon        APIPokemonItem                   `json:"pokemon"`
	VersionDetails []APIPokemonEncounterVersionItem `json:"version_details"`
}

type APIPokemonEncounterVersionItem struct {
	Version          APIVersionItem            `json:"version"`
	MaxChance        int                       `json:"max_chance"`
	EncounterDetails []APIEncounterDetailsItem `json:"encounter_details"`
}

type APIVersionDetailsItem struct {
	Rate    int            `json:"rate"`
	Version APIVersionItem `json:"version"`
}

type APIVersionItem struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type APIEncounterDetailsItem struct {
	MinLevel        int                    `json:"min_level"`
	MaxLevel        int                    `json:"max_level"`
	ConditionValues []string               `json:"condition_values"`
	Chance          int                    `json:"chance"`
	Method          APIEncounterMethodItem `json:"method"`
}

type APIEncounterMethodItem struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type APIPokemonByNameResponseBody struct {
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
	Height         int    `json:"height"`
	Weight         int    `json:"weight"`
	Stats          []struct {
		BaseStat int `json:"base_stat"`
		Effort   int `json:"effort"`
		Stat     struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Slot int `json:"slot"`
		Type struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"type"`
	} `json:"types"`
}

func Init(offset int) PokeapiClient {
	initialNext := fmt.Sprintf("https://pokeapi.co/api/v2/ability/?limit=20&offset=%d", offset)
	cache := pokecache.NewCache[[]byte](5 * time.Minute)
	client := PokeapiClient{
		prev:  nil,
		next:  &initialNext,
		cache: *cache,
	}
	return client
}

func updatePageLinks(client *PokeapiClient, body APIPageResponseBody) {
	if body.Next == "" {
		client.next = nil
	} else {
		client.next = &body.Next
	}
	if body.Prev == "" {
		client.prev = nil
	} else {
		client.prev = &body.Prev
	}
}

func get(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode > 299 {
		return nil, fmt.Errorf("Response failed with status code: %d and \nbody: %s\n", res.StatusCode, body)
	}
	if err != nil {
		return nil, err
	}
	return body, nil
}
