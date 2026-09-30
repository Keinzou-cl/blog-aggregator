package main

import (
	"log"
	"os"

	"database/sql"

	"github.com/Keinzou-cl/blog-aggregator/internal/config"
	"github.com/Keinzou-cl/blog-aggregator/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Error reading config: %v", err)
	}
	appState := &state{cfg: cfg}
	db, err := sql.Open("postgres", cfg.DBUrl)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	appState.db = database.New(db)
	commands := &commands{cmnds: make(map[string]func(*state, command) error)}
	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)
	commands.register("reset", handlerReset)
	commands.register("users", handlerUsers)
	commands.register("agg", handlerAggregate)
	commands.register("addfeed", middlewareLoggedIn(AddFeed))
	commands.register("feeds", handlerFeeds)
	commands.register("follow", middlewareLoggedIn(handlerFollow))
	commands.register("following", middlewareLoggedIn(handlerFollowing))
	commands.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	commands.register("browse", middlewareLoggedIn(handlerBrowse))
	args := os.Args[1:]
	if len(args) == 0 {
		log.Fatalf("No command provided")
	}
	cmd := command{name: args[0], args: args[1:]}
	if err := commands.run(appState, cmd); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}

}
