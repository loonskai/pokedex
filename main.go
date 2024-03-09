package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	pokeapi "github.com/loonskai/pokedexcli/internal/pokeapi"
)

type cliCommand struct {
	name        string
	description string
	callback    func(client *pokeapi.PokeapiClient, args ...string) error
}

type mapConfig struct {
	prev *string
	next *string
}

func main() {
	commandsMap := getCommadsMap()
	pokeapiClient := pokeapi.Init(0)
	var input *bufio.Scanner
	for {
		fmt.Print("Pokedex > ")
		input = bufio.NewScanner(os.Stdin)
		input.Scan()
		inputValues := strings.Fields(input.Text())
		inputCommand := inputValues[0]
		inputArgs := inputValues[1:]
		if command, ok := commandsMap[inputCommand]; ok {
			if err := command.callback(&pokeapiClient, inputArgs...); err != nil {
				fmt.Printf("Bad command execution: %v\n", err)
			}
		} else {
			fmt.Printf("Unknown command: %s\n", inputCommand)
		}
	}
}

func getCommadsMap() map[string]cliCommand {
	return map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help messsage",
			callback:    commandHelp,
		},
		"exit": {
			name:        "name",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"map": {
			name:        "map",
			description: "Get the next Pokemons batch",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Get the previous Pokemons batch",
			callback:    commandMapb,
		},
		"explore": {
			name:        "explore",
			description: "Explore pokemons in the area",
			callback:    commandExplore,
		},
	}
}

func commandHelp(client *pokeapi.PokeapiClient, args ...string) error {
	commandsMap := getCommadsMap()
	fmt.Println("Welcome to the Pokedex!\n\nUsage:")
	for key, command := range commandsMap {
		fmt.Printf("%s: %s\n", key, command.description)
	}
	return nil
}

func commandExit(client *pokeapi.PokeapiClient, args ...string) error {
	os.Exit(0)
	return nil
}

func commandMap(client *pokeapi.PokeapiClient, args ...string) error {
	pokemons, err := client.GetNext()
	if err != nil {
		return err
	}
	for _, pokemon := range pokemons.Results {
		fmt.Println(pokemon.Name)
	}
	return nil
}

func commandMapb(client *pokeapi.PokeapiClient, args ...string) error {
	pokemons, err := client.GetPrev()
	if err != nil {
		return err
	}
	for _, pokemon := range pokemons.Results {
		fmt.Println(pokemon.Name)
	}
	return nil
}

func commandExplore(client *pokeapi.PokeapiClient, args ...string) error {
	location := args[0]
	fmt.Printf("Exploring %s...\n", location)
	pokemons, err := client.GetFromLocationAreas(location)
	if err != nil {
		return err
	}
	for _, pokemonEncounter := range pokemons.PokemonEncounters {
		fmt.Printf(" - %s\n", pokemonEncounter.Pokemon.Name)
	}
	return nil
}
