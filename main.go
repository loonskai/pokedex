package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"

	pokeapi "github.com/loonskai/pokedexcli/internal/pokeapi"
)

type cliCommand struct {
	name        string
	description string
	callback    func(client *pokeapi.PokeapiClient, pokemons *map[string]pokemon, args ...string) error
}

type mapConfig struct {
	prev *string
	next *string
}

type pokemon struct {
	Name           string
	BaseExperience int
	Height         int
	Weight         int
	Stats          map[string]int
	Types          []string
}

func main() {
	commandsMap := getCommadsMap()
	pokeapiClient := pokeapi.Init(0)
	caughtPokemons := map[string]pokemon{}
	var input *bufio.Scanner
	for {
		fmt.Print("Pokedex > ")
		input = bufio.NewScanner(os.Stdin)
		input.Scan()
		inputValues := strings.Fields(input.Text())
		inputCommand := inputValues[0]
		inputArgs := inputValues[1:]
		if command, ok := commandsMap[inputCommand]; ok {
			if err := command.callback(&pokeapiClient, &caughtPokemons, inputArgs...); err != nil {
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
		"catch": {
			name:        "catch",
			description: "Catch a pokemon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Inspect a pokemon",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "List all caught pokemons",
			callback:    commandPokedex,
		},
	}
}

func commandHelp(client *pokeapi.PokeapiClient, caughtPokemons *map[string]pokemon, args ...string) error {
	commandsMap := getCommadsMap()
	fmt.Println("Welcome to the Pokedex!\n\nUsage:")
	for key, command := range commandsMap {
		fmt.Printf("%s: %s\n", key, command.description)
	}
	return nil
}

func commandExit(client *pokeapi.PokeapiClient, caughtPokemons *map[string]pokemon, args ...string) error {
	os.Exit(0)
	return nil
}

func commandMap(client *pokeapi.PokeapiClient, caughtPokemons *map[string]pokemon, args ...string) error {
	pokemons, err := client.GetNext()
	if err != nil {
		return err
	}
	for _, pokemon := range pokemons.Results {
		fmt.Println(pokemon.Name)
	}
	return nil
}

func commandMapb(client *pokeapi.PokeapiClient, caughtPokemons *map[string]pokemon, args ...string) error {
	pokemons, err := client.GetPrev()
	if err != nil {
		return err
	}
	for _, pokemon := range pokemons.Results {
		fmt.Println(pokemon.Name)
	}
	return nil
}

func commandExplore(client *pokeapi.PokeapiClient, caughtPokemons *map[string]pokemon, args ...string) error {
	location := args[0]
	if location == "" {
		return fmt.Errorf("Missing location")
	}
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

func commandCatch(client *pokeapi.PokeapiClient, caughtPokemons *map[string]pokemon, args ...string) error {
	name := args[0]
	if name == "" {
		return fmt.Errorf("Missing name")
	}
	fmt.Printf("Throwing a Pokeball at %s...\n", name)
	pokemonFound, err := client.GetPokemonByName(name)
	if err != nil {
		return err
	}
	threshold := 50
	randNum := rand.Intn(pokemonFound.BaseExperience)
	if randNum > threshold {
		return fmt.Errorf("Failed to catch %s\n", name)
	}
	caught := pokemon{
		Name:           pokemonFound.Name,
		BaseExperience: pokemonFound.BaseExperience,
		Height:         pokemonFound.Height,
		Weight:         pokemonFound.Weight,
		Stats:          map[string]int{},
		Types:          []string{},
	}
	for _, stat := range pokemonFound.Stats {
		caught.Stats[stat.Stat.Name] = stat.BaseStat
	}
	for _, t := range pokemonFound.Types {
		caught.Types = append(caught.Types, t.Type.Name)
	}
	(*caughtPokemons)[name] = caught
	fmt.Printf("%s was caught!\n", name)
	fmt.Println("You may now inspect it with the inspect command.")
	return nil
}

func commandInspect(client *pokeapi.PokeapiClient, caughtPokemons *map[string]pokemon, args ...string) error {
	name := args[0]
	if name == "" {
		return fmt.Errorf("Missing name")
	}
	caught, ok := (*caughtPokemons)[name]
	if !ok {
		return fmt.Errorf("You have not caught that pokemon")
	}
	fmt.Printf("Name: %s\n", caught.Name)
	fmt.Printf("Height: %s\n", caught.Height)
	fmt.Printf("Weight: %s\n", caught.Weight)
	fmt.Println("Stats:")
	for key, val := range caught.Stats {
		fmt.Printf(" -%s: %d\n", key, val)
	}
	fmt.Println("Types:")
	for _, t := range caught.Types {
		fmt.Printf(" - %s\n", t)
	}
	return nil
}

func commandPokedex(client *pokeapi.PokeapiClient, caughtPokemons *map[string]pokemon, args ...string) error {
	fmt.Println("Your Pokedex:")
	if len(*caughtPokemons) == 0 {
		fmt.Println("Empty :(")
	} else {
		for key := range *caughtPokemons {
			fmt.Printf(" - %s\n", key)
		}
	}
	return nil
}
