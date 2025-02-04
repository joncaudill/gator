package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/joncaudill/gator/internal/database"
)

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

	err = s.db.ResetFeedFollows(context.Background())
	if err != nil {
		fmt.Printf("could not reset feed follows: %s", err)
		os.Exit(1)
	}
	fmt.Println("feed follows table was reset")

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

func handlerAgg(s *state, cmd command) error {
	//func that aggregates the RSS feeds

	//the fake feed URL is used for testing
	//remove in production
	//cmd.args = []string{"https://www.wagslane.dev/index.xml"}

	if len(cmd.args) == 0 {
		return fmt.Errorf("agg command requires 1 or more arguments")
	}

	time_between_reqs := cmd.args[0]
	ticker_duration, err := time.ParseDuration(time_between_reqs)
	ticker_min, _ := time.ParseDuration("10m")
	if ticker_duration < ticker_min {
		ticker_duration = ticker_min
	}
	if err != nil {
		return fmt.Errorf("could not parse time duration: %w", err)
	}

	fmt.Printf("Collecting feeds every %s\n", ticker_duration)

	//do an initial scrape of the feeds	before starting the ticker
	err = scrapeFeeds(s)
	if err != nil {
		return fmt.Errorf("could not scrape feeds: %w", err)
	}

	ticker := time.NewTicker(ticker_duration)
	for ; ; <-ticker.C {
		err := scrapeFeeds(s)
		if err != nil {
			return fmt.Errorf("could not scrape feeds: %w", err)
		}
	}

	// old code used to test feed aggregation -- TODO remove eventually
	// for _, feedURL := range cmd.args {
	// 	feed, err := fetchFeed(context.Background(), feedURL)
	// 	if err != nil {
	// 		fmt.Printf("could not fetch feed: %s", err)
	// 		os.Exit(1)
	// 	}
	// 	fmt.Printf("Feed: %s\n", feed.Channel.Title)
	// 	fmt.Printf("Link: %s\n", feed.Channel.Link[0])
	// 	fmt.Printf("Description: %s\n", feed.Channel.Description)
	// 	fmt.Println("Items:")
	// 	for _, item := range feed.Channel.Item {
	// 		fmt.Printf("  * %s\n", item.Title)
	// 		fmt.Printf("    %s\n", item.Link)
	// 		fmt.Printf("    %s\n", item.Description)
	// 		fmt.Printf("    %s\n", item.PubDate)
	// 	}
	// }
	// return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	//func that adds a feed to the feeds table
	if len(cmd.args) != 2 {
		fmt.Println("addfeed command requires 2 arguments")
		os.Exit(1)
	}

	feedid := uuid.New()
	timeNow := time.Now()

	feed, err := s.db.CreateFeed(context.Background(),
		database.CreateFeedParams{ID: uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      cmd.args[0],
			Url:       cmd.args[1],
			UserID:    user.ID,
		})

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

	_, err = s.db.CreateFeedFollow(context.Background(),
		database.CreateFeedFollowParams{ID: uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			UserID:    user.ID,
			FeedID:    feed.ID,
		})

	if err != nil {
		fmt.Printf("could not create feed follow: %s", err)
		os.Exit(1)
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

func handlerAddFollow(s *state, cmd command, user database.User) error {
	//func that adds a follow to the feed follows table
	if len(cmd.args) == 0 {
		fmt.Println("addfollow command requires 1 argument")
		os.Exit(1)
	}

	feed, err := s.db.GetFeedByUrl(context.Background(), cmd.args[0])
	if err != nil {
		fmt.Printf("could not get feed by URL: %s", err)
		os.Exit(1)
	}

	timeNow := time.Now()
	followid := uuid.New()

	_, err = s.db.CreateFeedFollow(context.Background(),
		database.CreateFeedFollowParams{ID: followid,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
			UserID:    user.ID,
			FeedID:    feed.ID,
		})

	if err != nil {
		fmt.Printf("could not create feed follow: %s", err)
		os.Exit(1)
	}

	return nil
}

func handlerDeleteFollow(s *state, cmd command, user database.User) error {
	//func that deletes a follow from the feed follows table
	if len(cmd.args) == 0 {
		fmt.Println("deletefollow command requires 1 argument")
		os.Exit(1)
	}

	feed, err := s.db.GetFeedByUrl(context.Background(), cmd.args[0])
	if err != nil {
		fmt.Printf("could not get feed by URL: %s", err)
		os.Exit(1)
	}

	err = s.db.DeleteFeedFollow(context.Background(),
		database.DeleteFeedFollowParams{UserID: user.ID,
			FeedID: feed.ID,
		})

	if err != nil {
		fmt.Printf("could not delete feed follow: %s", err)
		os.Exit(1)
	}

	fmt.Printf("Unfollowed feed: %s\n", feed.Name)

	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	//func that lists all the feeds that the current user is following

	follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		fmt.Printf("could not get feed follows by user: %s", err)
		os.Exit(1)
	}

	if len(follows) == 0 {
		fmt.Println("No feeds are being followed.")
		return nil
	}

	fmt.Println("Following:")
	for _, follow := range follows {
		fmt.Printf("* %s\n", follow.FeedName)
	}

	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	//func that takes a limit parameter and lists all the posts in the posts table
	//that the current user is following, limited by the limit parameter

	if len(cmd.args) == 0 {
		cmd.args = append(cmd.args, "2")
	}

	limit, err := strconv.Atoi(cmd.args[0])
	if err != nil {
		limit = 2
	}

	posts, err := s.db.GetPostsForUser(context.Background(),
		database.GetPostsForUserParams{UserID: user.ID,
			Limit: int32(limit),
		})

	if err != nil {
		fmt.Printf("could not get posts for user: %s", err)
		os.Exit(1)
	}

	if len(posts) == 0 {
		fmt.Println("No posts to display.")
		return nil
	}

	fmt.Println("Posts:")
	for _, post := range posts {
		fmt.Printf("* %s\n", post.Title)
		fmt.Printf("  %s\n", post.Url)
		fmt.Printf("  %s\n", post.Description)
		fmt.Printf("  %s\n", post.PublishedAt)
	}

	return nil
}
