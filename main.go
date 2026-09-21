package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

type config struct {
	commands   map[string]cliCommand
	Next       *string
	Previous   *string
	httpClient *http.Client
}

type cliCommand struct {
	name        string
	description string
	callback    func(*config) error
}

func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"map":  {name: "map", description: "Display the next 20 location areas", callback: commandMap},
		"mapb": {name: "mapb", description: "Display the previous 20 location areas", callback: commandMapb},
		"exit": {name: "exit", description: "Exit the Pokedex", callback: commandExit},
		"help": {name: "help", description: "Displays a help message", callback: commandHelp},
	}
}

func commandExit(_ *config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *config) error {
	fmt.Print("Welcome to the Pokedex!\nUsage:\n\n")
	commands := cfg.commands
	names := make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}
	// Show help first, then keep the remaining commands in alphabetical order.
	sort.Slice(names, func(i, j int) bool {
		if names[i] == "help" || names[j] == "help" {
			return names[i] == "help"
		}
		return names[i] < names[j]
	})
	for _, name := range names {
		command := commands[name]
		fmt.Printf("%s: %s\n", command.name, command.description)
	}
	return nil
}

func cleanInput(text string) []string {
	return strings.Fields(strings.ToLower(text))
}

func main() {
	firstPage := "https://pokeapi.co/api/v2/location-area/?limit=20"
	cfg := &config{
		commands:   getCommands(),
		Next:       &firstPage,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
	startRepl(cfg)
}

func startRepl(cfg *config) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		if !scanner.Scan() {
			break
		}

		words := cleanInput(scanner.Text())
		if len(words) == 0 {
			continue
		}
		command, ok := cfg.commands[words[0]]
		if !ok {
			fmt.Println("Unknown command")
			continue
		}
		if err := command.callback(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
	}
}
