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
	//func that resets the user table and feeds table
	//this is a dangerous command and should not be used in production
	err := s.db.ResetUsers(context.Background())
	if err != nil {
		fmt.Printf("could not reset users: %s", err)
		os.Exit(1)
	}
	fmt.Println("users table was reset")

	err = s.db.ResetFeeds(context.Background())
	if err != nil {
		fmt.Printf("could not reset feeds: %s", err)
		os.Exit(1)
	}
	fmt.Println("feeds table was reset")

	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	//func that adds a feed to the feeds table
	if len(cmd.args) != 2 {
		fmt.Println("addfeed command requires 2 arguments")
		os.Exit(1)
	}

	//since this is tied to the current user, we need to get the current user
	user, err := getCurrentUser(s)
	if err != nil {
		fmt.Printf("could not get current user: %s", err)
		os.Exit(1)
	}

	feedid := uuid.New()
	timeNow := time.Now()

	_, err = s.db.CreateFeed(context.Background(),
		database.CreateFeedParams{ID: uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      cmd.args[0],
			Url:       cmd.args[1],
			UserID:    user.ID})

	if err != nil {
		fmt.Printf("could not create feed: %s", err)
		os.Exit(1)
	}

	//print the fields of the newly created feed
	fmt.Println("feed was created.")
	fmt.Printf("ID: %s\n", feedid)
	fmt.Printf("Created At: %s\n", timeNow)
	fmt.Printf("Updated At: %s\n", timeNow)
	fmt.Printf("Name: %s\n", cmd.args[0])
	fmt.Printf("URL: %s\n", cmd.args[1])
	fmt.Printf("User ID: %s\n", user.ID)

	return nil
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

func handlerAgg(s *state, cmd command) error {
	//func that aggregates the RSS feeds

	//the fake feed URL is used for testing
	//remove in production
	cmd.args = []string{"https://www.wagslane.dev/index.xml"}

	if len(cmd.args) == 0 {
		return fmt.Errorf("agg command requires 1 or more arguments")
	}

	for _, feedURL := range cmd.args {
		feed, err := fetchFeed(context.Background(), feedURL)
		if err != nil {
			fmt.Printf("could not fetch feed: %s", err)
			os.Exit(1)
		}
		fmt.Printf("Feed: %s\n", feed.Channel.Title)
		fmt.Printf("Link: %s\n", feed.Channel.Link[0])
		fmt.Printf("Description: %s\n", feed.Channel.Description)
		fmt.Println("Items:")
		for _, item := range feed.Channel.Item {
			fmt.Printf("  * %s\n", item.Title)
			fmt.Printf("    %s\n", item.Link)
			fmt.Printf("    %s\n", item.Description)
			fmt.Printf("    %s\n", item.PubDate)
		}
	}
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	//func that lists all the feeds in the feeds table
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		fmt.Printf("could not get feeds: %s", err)
		os.Exit(1)
	}
	for _, feed := range feeds {
		fmt.Printf("*Feed Name: %s\n", feed.Name)
		fmt.Printf("Feed URL:  %s\n", feed.Url)
		feedUser, err := getUserById(s, feed.UserID)
		if err != nil {
			fmt.Printf("could not get user by id: %s", err)
			os.Exit(1)
		}
		fmt.Printf("Created By: %s\n", feedUser.Name)
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
	cliCommands.register("agg", handlerAgg)
	cliCommands.register("addfeed", handlerAddFeed)
	cliCommands.register("feeds", handlerFeeds)

	args := os.Args
	if len(args) < 2 {
		fmt.Println("not enough arguments were provided")
		os.Exit(1)
	}

	cliCommands.run(cliState, command{name: args[1], args: args[2:]})

}
