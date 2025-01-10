package main

import (
	"context"
	"database/sql"
	"fmt"
	"internal/config"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joncaudill/gator/internal/database"
	_ "github.com/lib/pq"
)

type state struct {
	//struct that represents the state of the application
	db     *database.Queries
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

	_, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err != nil {
		fmt.Printf("could not get user: %s", err)
		os.Exit(1)
	}
	err = s.config.SetUser(cmd.args[0])
	if err != nil {
		return fmt.Errorf("could not set user name: %w", err)
	}
	fmt.Println("user name set to:", cmd.args[0])
	return nil
}

func handlerRegister(s *state, cmd command) error {
	//func that handles the register command
	if len(cmd.args) == 0 {
		return fmt.Errorf("register command requires 1 argument")
	}

	user, err := s.db.CreateUser(context.Background(),
		database.CreateUserParams{ID: uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      cmd.args[0]})

	if err != nil {
		fmt.Printf("could not create user: %s", err)
		os.Exit(1)
	}
	s.config.SetUser(user.Name)
	fmt.Println("user was created.")
	fmt.Printf("user id: %s\n", user.ID)
	fmt.Printf("user name: %s\n", user.Name)
	fmt.Printf("created at: %s\n", user.CreatedAt)
	fmt.Printf("updated at: %s\n", user.UpdatedAt)

	return nil
}

func handlerReset(s *state, cmd command) error {
	//func that resets the user table
	//this is a dangerous command and should not be used in production
	err := s.db.ResetUsers(context.Background())
	if err != nil {
		fmt.Printf("could not reset users: %s", err)
		os.Exit(1)
	}
	fmt.Println("users table was reset")
	return nil
}

func handlerList(s *state, cmd command) error {
	//func that lists all the users in the user table
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		fmt.Printf("could not get users: %s", err)
		os.Exit(1)
	}
	for _, user := range users {
		status := ""
		if user.Name == s.config.UserName {
			status = " (current)"
		}
		fmt.Printf("* %s%s\n", user.Name, status)
	}
	return nil
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println("could not read config:", err)
		return
	}

	db, err := sql.Open("postgres", cfg.DbUrl)
	if err != nil {
		fmt.Println("could not connect to database:", err)
		return
	}
	dbQueries := database.New(db)

	cliState := &state{config: &cfg, db: dbQueries}
	cliCommands := commands{names: make(map[string]func(*state, command) error)}
	cliCommands.register("login", handlerLogin)
	cliCommands.register("register", handlerRegister)
	cliCommands.register("reset", handlerReset)
	cliCommands.register("users", handlerList)

	args := os.Args
	if len(args) < 2 {
		fmt.Println("not enough arguments were provided")
		os.Exit(1)
	}

	cliCommands.run(cliState, command{name: args[1], args: args[2:]})

}
