package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"internal/config"
	"io"
	"net/http"
	"os"

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

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        []string  `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
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

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	//fetches a given RSS feed from a URL
	request, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	//set the request headers
	request.Header.Set("User-Agent", "gator-cli")

	//use client with the context
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("could not fetch feed: %w", err)
	}
	defer response.Body.Close()

	//parse the response body
	var feed RSSFeed
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %w", err)
	}
	err = xml.Unmarshal(body, &feed)
	if err != nil {
		return nil, fmt.Errorf("could not parse feed: %w", err)
	}

	//unescape the HTML entities in the feed
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	for i := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
		feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
	}

	return &feed, nil
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(s *state, cmd command) error {
	//middleware that checks if the user is logged in
	return func(s *state, cmd command) error {
		user, err := getCurrentUser(s)
		if err != nil {
			return fmt.Errorf("could not get current user: %w", err)
		}
		return handler(s, cmd, user)
	}
}

func getCurrentUser(s *state) (database.User, error) {
	//func that gets the current user name
	user, err := s.db.GetUser(context.Background(), s.config.UserName)
	if err != nil {
		return database.User{}, fmt.Errorf("could not get user name: %w", err)
	}
	return user, nil
}

func getUserById(s *state, id uuid.UUID) (database.User, error) {
	//func that gets the user by the user id
	user, err := s.db.GetUserById(context.Background(), id)
	if err != nil {
		return database.User{}, fmt.Errorf("could not get user by id: %w", err)
	}
	return user, nil
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
	cliCommands.register("agg", handlerAgg)
	cliCommands.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cliCommands.register("feeds", handlerFeeds)
	cliCommands.register("follow", middlewareLoggedIn(handlerAddFollow))
	cliCommands.register("following", middlewareLoggedIn(handlerFollowing))

	args := os.Args
	if len(args) < 2 {
		fmt.Println("not enough arguments were provided")
		os.Exit(1)
	}

	cliCommands.run(cliState, command{name: args[1], args: args[2:]})

}
