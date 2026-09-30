package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"database/sql"

	"github.com/Keinzou-cl/blog-aggregator/internal/config"
	"github.com/Keinzou-cl/blog-aggregator/internal/database"
	"github.com/Keinzou-cl/blog-aggregator/internal/rss"
	"github.com/google/uuid"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	cmnds map[string]func(*state, command) error
}

func handlerLogin(s *state, cmd command) error {

	if len(cmd.args) == 0 {
		return fmt.Errorf("username is required")
	}
	username := cmd.args[0]

	_, err := s.db.GetUser(context.Background(), username)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if err := s.cfg.SetUser(username); err != nil {
		return fmt.Errorf("failed to set user: %w", err)
	}

	fmt.Printf("Current user set to: %s\n", s.cfg.CurrentUserName)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("username is required")
	}
	username := cmd.args[0]
	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      username,
	})
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	if err := s.cfg.SetUser(username); err != nil {
		return fmt.Errorf("failed to set user: %w", err)
	}
	fmt.Printf("User created: %s (ID: %s)\n", user.Name, user.ID)
	return nil
}

func handlerReset(s *state, cmd command) error {
	if err := s.db.DeleteAllUsers(context.Background()); err != nil {
		return fmt.Errorf("failed to delete all users: %w", err)
	}
	fmt.Println("All users deleted.")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get users: %w", err)
	}
	for _, user := range users {
		if user.Name == s.cfg.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

func handlerAggregate(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("feed URL is required")
	}
	timeBetweenRequests, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}
	fmt.Println("Collecting feeds every", timeBetweenRequests)
	ticker := time.NewTicker(timeBetweenRequests)
	defer ticker.Stop()
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func AddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("name and url are required")
	}
	name := cmd.args[0]
	url := cmd.args[1]

	ctx := context.Background()

	feed, err := s.db.CreateFeed(ctx, database.CreateFeedParams{
		ID:        uuid.New(),
		Name:      name,
		Url:       url,
		UserID:    user.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return fmt.Errorf("failed to create feed: %w", err)
	}

	followFeed, err := s.db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID:        uuid.New(),
		UserID:    user.ID,
		FeedID:    feed.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return fmt.Errorf("failed to create feed follow: %w", err)
	}

	fmt.Printf("Feed created: %s (ID: %s)\n", feed.Name, feed.ID)
	fmt.Printf("User %s is now following feed %s\n", followFeed.UserName, followFeed.FeedName)

	feedData, err := rss.FetchFeed(ctx, url)
	if err != nil {
		return fmt.Errorf("failed to fetch feed data: %w", err)
	}

	fmt.Printf("Feed: %+v\n", feedData)
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	ctx := context.Background()

	feed, err := s.db.GetFeeds(ctx)
	if err != nil {
		return fmt.Errorf("failed to get feeds: %w", err)
	}

	for _, f := range feed {
		fmt.Printf("Feed Name: %s\n", f.Name)
		fmt.Printf("URL: %s\n", f.Url)
		fmt.Printf("User Name: %s\n", f.UserName)
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("feed name is required")
	}

	ctx := context.Background()

	url := cmd.args[0]
	feed, err := s.db.GetFeedbyUrl(ctx, url)
	if err != nil {
		return fmt.Errorf("failed to get feed by URL: %w", err)
	}

	feedFollow, err := s.db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID:        uuid.New(),
		UserID:    user.ID,
		FeedID:    feed.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	if err != nil {
		return fmt.Errorf("failed to create follow: %w", err)
	}

	fmt.Printf("User %s is now following feed %s\n", feedFollow.FeedName, feedFollow.UserName)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	ctx := context.Background()
	following, err := s.db.GetFeedFollowsForUser(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("failed to get following feeds: %w", err)
	}

	for _, f := range following {
		fmt.Printf("Feed Name: %s\n", f.FeedName)
		fmt.Printf("User Name: %s\n", f.UserName)
	}
	return nil
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		ctx := context.Background()
		user, err := s.db.GetUser(ctx, s.cfg.CurrentUserName)
		if err != nil {
			return fmt.Errorf("failed to get current user: %w", err)
		}
		return handler(s, cmd, user)
	}
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("feed name is required")
	}

	ctx := context.Background()

	url := cmd.args[0]
	feed, err := s.db.GetFeedbyUrl(ctx, url)
	if err != nil {
		return fmt.Errorf("failed to get feed by URL: %w", err)
	}

	err = s.db.DeleteFeedFollow(ctx, database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to unfollow feed: %w", err)
	}

	fmt.Printf("User %s has unfollowed feed %s\n", user.Name, feed.Name)
	return nil
}

func scrapeFeeds(s *state) error {
	ctx := context.Background()
	feed, err := s.db.GetNextFeedToFetch(ctx)
	if err != nil {
		return fmt.Errorf("failed to get next feed to fetch: %w", err)
	}

	feedData, err := rss.FetchFeed(ctx, feed.Url)
	if err != nil {
		return fmt.Errorf("failed to fetch feed data: %w", err)
	}

	fmt.Printf("Fetched Feed: %+v\n", feedData)

	err = s.db.MarkFeedFetched(ctx, feed.ID)
	if err != nil {
		return fmt.Errorf("failed to mark feed as fetched: %w", err)
	}

	for _, item := range feedData.Channel.Item {
		publishedAt, err := time.Parse(time.RFC1123, item.PubDate)
		if err != nil {
			return fmt.Errorf("failed to parse publication date: %w", err)
		}
		_, err = s.db.CreatePost(ctx, database.CreatePostParams{
			ID:          uuid.New(),
			FeedID:      feed.ID,
			Title:       item.Title,
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: true},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			PublishedAt: publishedAt,
		})
		if err != nil {
			return fmt.Errorf("failed to create post: %w", err)
		}
	}

	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	ctx := context.Background()
	limit := 2
	if len(cmd.args) > 0 {
		num, err := strconv.Atoi(cmd.args[0])
		if err != nil {
			return fmt.Errorf("invalid limit: %w", err)
		}
		limit = num
	}

	posts, err := s.db.GetPostsForUser(ctx, database.GetPostsForUserParams{
		Name:  user.Name,
		Limit: int32(limit),
	})
	if err != nil {
		return fmt.Errorf("failed to get posts: %w", err)
	}

	for _, post := range posts {
		fmt.Printf("Post Title: %s\n", post.Title)
		fmt.Printf("URL: %s\n", post.Url)
		fmt.Printf("Description: %s\n", post.Description.String)
		fmt.Printf("Published At: %s\n", post.PublishedAt.Format(time.RFC1123))
		fmt.Println("---")
	}
	return nil
}

func (c *commands) run(s *state, cmd command) error {
	if handler, exists := c.cmnds[cmd.name]; exists {
		return handler(s, cmd)
	}
	return fmt.Errorf("unknown command: %s", cmd.name)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.cmnds[name] = f
}
