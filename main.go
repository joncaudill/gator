package main

import (
	"fmt"
	"internal/config"
	"os"
)

type state struct {
	//struct that represents the state of the application
	config *config.Config
}

type command struct {
	//struct that contains the params of the command
	name string
	args []string
}

type commands struct {
	//struct that contains the commands of the application
	names map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	//registers a new handler function for a command name
	c.names[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	//runs a given command with the state passed into the func
	err := c.names[cmd.name](s, cmd)
	if err != nil {
		return fmt.Errorf("could not run command: %w", err)
	}
	return nil
}

func handlerLogin(s *state, cmd command) error {
	//func that handles the login command
	if len(cmd.args) == 0 {
		return fmt.Errorf("login command requires 1 argument")
	}
	err := s.config.SetUser(cmd.args[0])
	if err != nil {
		return fmt.Errorf("could not set user name: %w", err)
	}
	fmt.Println("user name set to:", cmd.args[0])
	return nil
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println("could not read config:", err)
		return
	}
	cliState := &state{config: &cfg}
	cliCommands := commands{names: make(map[string]func(*state, command) error)}
	cliCommands.register("login", handlerLogin)

	args := os.Args
	if len(args) < 2 {
		fmt.Println("not enough arguments were provided")
		os.Exit(1)
	}
	if len(args) < 3 {
		fmt.Println("a username is required")
		os.Exit(1)
	}
	cliCommands.run(cliState, command{name: args[1], args: args[2:]})

}
