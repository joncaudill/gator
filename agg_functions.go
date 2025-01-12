package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joncaudill/gator/internal/database"
)

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

func scrapeFeeds(s *state) error {
	//scrapeFeeds fetches the next feed to fetch from the database
	//using GetFeedToFetch query and then fetches the feed
	//afterward it marks the feed as fetched using the MarkFeedFetched query
	//and then iterates over all the items in the feed and prints their titles to the console
	feed, err := s.db.GetFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("could not get feed to fetch: %w", err)
	}

	err = s.db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		return fmt.Errorf("could not mark feed fetched: %w", err)
	}

	feedURL := feed.Url
	feedRSS, err := fetchFeed(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("could not fetch feed: %w", err)
	}

	for _, item := range feedRSS.Channel.Item {
		//fmt.Printf("  * %s\n", item.Title)
		publishedAt, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			return fmt.Errorf("could not parse time: %w", err)
		}
		if item.Title == "" {
			item.Title = "No Title"
		}
		_, err = s.db.CreatePost(context.Background(),
			database.CreatePostParams{ID: uuid.New(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				Title:       item.Title,
				Url:         item.Link,
				Description: item.Description,
				PublishedAt: publishedAt,
				FeedID:      feed.ID,
			})
		if err != nil {
			if !strings.Contains(err.Error(), "duplicate key") {
				fmt.Printf("could not create feed item: %s", err)
			}
		}

	}

	return nil

}
