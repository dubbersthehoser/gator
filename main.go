package main

import (
	
	"log"
	"fmt"
	"os"
	"database/sql"
	"context"
	"time"
	"net/http"
	"encoding/xml"
	"html"
	"io"
	"bytes"
	"strconv"
	
	"github.com/dubbersthehoser/gator/internal/config"
	"github.com/dubbersthehoser/gator/internal/database"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)


type state struct {
	config *config.Config
	db     *database.Queries
	cmds   commands
}

type command struct {
	Name string
	Args []string
}

type commands struct {
	Map map[string]func(s *state, cmd command) error
}

func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.Map[cmd.Name]
	if !ok {
		return fmt.Errorf("commands: %s, command not founnd", cmd.Name)
	}
	return handler(s, cmd)
}
func (c *commands) register(name string, f func(*state, command) error) {
	c.Map[name] = f
}

//
// Middleware
//
func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		if s.config.CurrentUserName == nil {
			return fmt.Errorf("CurrentUserName is nil")
		}

		user, err := s.db.GetUser(context.Background(), *s.config.CurrentUserName)	
		if err != nil {
			return err
		}
		handler(s, cmd, user)
		return nil
	}
}

//
// Aggergate
//
func scrapeFeeds(s *state) error {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}
	err = s.db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		return err
	}
	rss, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		return err
	}

	postParams := database.CreatePostParams{}
	for _, item := range rss.Channel.Items {
		if item.Title == "" || item.Link == "" {
			continue
		}
		description := sql.NullString{String: "", Valid: false}
		if item.Description == "" {
			description = sql.NullString{String: item.Description, Valid: true}
		}

		pubDate := sql.NullTime{Valid:false}
		t, err := time.Parse(time.RFC822, item.PubDate)
		if err != nil {
			pubDate = sql.NullTime{Time: t, Valid: true}
		}

		id, err := s.db.GetPostsCount(context.Background())
		if err != nil {
			return err
		}

		postParams.ID = id
		postParams.CreatedAt = time.Now()
		postParams.UpdatedAt = time.Now()
		postParams.Title = item.Title
		postParams.Url = item.Link
		postParams.Description = description
		postParams.PublishedAt = pubDate
		postParams.FeedID = feed.ID

		s.db.CreatePost(context.Background(), postParams)
	}
	return nil
}


//
// Command Handlers
//
// Help Handler 
func handlerHelp(s *state, cmd command) error {
	for k, _ := range s.cmds.Map {
		fmt.Printf("* %s\n", k)
	}
	return nil
}
// Users Handler
func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("no username was given for login")
	}

	name := cmd.Args[0]

	_, err := s.db.GetUser(context.Background(), name)

	if err != nil {
		return err
	}
	
	s.config.SetUser(cmd.Args[0])
	fmt.Println("username set:", cmd.Args[0])
	return nil
}
func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("no name was given")
	}

	name := cmd.Args[0]
	id := uuid.New().String()
	pramUser := database.CreateUserParams{
		ID:        id,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
	}

	_, err := s.db.GetUser(context.Background(), name)
	if err == nil {
		return fmt.Errorf("already registered")
	}

	_, err = s.db.CreateUser(context.Background(), pramUser)
	if err != nil {
		return err
	}

	s.config.SetUser(name)

	fmt.Printf("%s user was created\n", name)
	fmt.Printf("DEBUG: %v\n", pramUser)

	return nil
}
func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetAllUsers(context.Background())
	if err != nil {
		return err
	}
	for _, user := range users {
		if user.Name == *s.config.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

// DB Handlers
func handlerReset(s *state, cmd command) error {
	err := s.db.DeleteAllUsers(context.Background())
	if err != nil {
		return err
	}
	err = s.db.DeleteAllFeeds(context.Background())
	if err != nil {
		return err
	}
	s.config.ClearUser()
	return nil
}

// Feed Handlers
func handlerBrowse(s *state, cmd command, user database.User) error {
	
	limit := 0
	if len(cmd.Args) == 0 {
		limit = 2
	} else if len(cmd.Args) != 1 {
		return fmt.Errorf("one or zero argument needed")
	} else {
		l, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("invalid limit argument: not an int")
		}
		limit = l
	}

	params := database.GetPostsForUserParams{
		Limit: int32(limit),
		ID: user.ID,
	}

	posts, err := s.db.GetPostsForUser(context.Background(), params)
	if err != nil {
		return err
	}

	for _, post := range posts {
		fmt.Printf("[%s]\n", post.Title)
		fmt.Printf("- Description: %s\n", post.Description)
		fmt.Printf("- link: %s\n", post.Url)
	}
	return nil
}
func handlerFeeds(s *state, cmd command) error {

	allFeeds, err := s.db.GetAllFeeds(context.Background())
	if err != nil {
		return err
	}

	usersFound := make(map[string]string)
	for _, feed := range allFeeds {
		userId := feed.UserID
		userName, ok := usersFound[userId]
		if !ok {
			user, err := s.db.GetUserByID(context.Background(), userId)
			if err != nil {
				return err
			}
			userName = user.Name
			usersFound[user.ID] = user.Name
		}

		fmt.Printf("%s: %s, %s\n", userName, feed.Name, feed.Url)
	}

	return nil
}
func handlerAgg(s *state, cmd command) error {

	if len(cmd.Args) != 1 {
		return fmt.Errorf("one argument needed: 'time'")
	}

	durration, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return err
	}

	if durration < time.Second {
		return fmt.Errorf("Error: durration time is smaller then a second")
	}

	ticker := time.NewTicker(durration)

	for ;; <-ticker.C {
		err := scrapeFeeds(s)
		fmt.Println(err)
	}
	return nil
}
func handlerFollowing(s *state, cmd command, user database.User) error {
	feeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.Name)
	if err != nil {
		return err
	}
	for _, feed := range feeds {
		fmt.Printf("%s\n", feed.FeedName)
	}
	return nil
}
func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("one argumant need: 'url'")
	}

	feedUrl := cmd.Args[0]

	feed, err := s.db.GetFeedByURL(context.Background(), feedUrl)
	if err != nil {
		return err
	}

	count, err := s.db.GetFeedFollowsCount(context.Background())
	if err != nil {
		return err
	}
	
	pramFeedFollow := database.CreateFeedFollowParams{
		ID: count,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: user.ID,
		FeedID: feed.ID,
	}

	newFeed, err := s.db.CreateFeedFollow(context.Background(), pramFeedFollow)
	if err != nil {
		return err
	}

	fmt.Printf("user='%s' feed='%s'\n", user.Name, newFeed.FeedName)
	return nil
}
func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("one argument needed: 'url'")
	}

	params := database.FeedUnfollowParams{
		Url: cmd.Args[0],
		Name: user.Name,
	}

	err := s.db.FeedUnfollow(context.Background(), params)
	if err != nil {
		return err
	}
	return nil

}
func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("two argument needed: 'name url'")
	}
	feedName := cmd.Args[0]
	feedUrl := cmd.Args[1]

	id, err := s.db.GetFeedCount(context.Background())
	if err != nil {
		return err
	}

	pramFeed := database.CreateFeedParams{
		ID:        id,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      feedName,
		Url:       feedUrl,
		UserID:    user.ID,
	}

	feed, err := s.db.CreateFeed(context.Background(), pramFeed)
	if err != nil {
		return err
	}

	fmt.Println("Feed added")
	fmt.Printf("DEBUG: %v\n", pramFeed)

	//
	// Make curr user follow feed
	//
	count, err := s.db.GetFeedFollowsCount(context.Background())
	if err != nil {
		return err
	}

	pramFeedFollow := database.CreateFeedFollowParams{
		ID: count,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: user.ID,
		FeedID: feed.ID,
	}

	_, err = s.db.CreateFeedFollow(context.Background(), pramFeedFollow)
	if err != nil {
		return err
	}

	return nil
}

//
// RSS
//
type RSSFeed struct {
	Channel struct {
		Title string `xml:"title"`
		Link  string `xml:"link"`
		Description string `xml:"description"`
		Items []RSSItem `xml:"item"`
	} `xml:"channel"`
}
type RSSItem struct {
	Title string `xml:"title"`
	Link  string `xml:"link"`
	Description string `xml:"description"`
	PubDate string `xml:"pubDate"`
}
func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, bytes.NewReader(make([]byte, 0)))
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", "gator")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	rss := RSSFeed{}
	err = xml.Unmarshal(data, &rss)
	if err != nil {
		return nil, err
	}

	rss.Channel.Title = html.UnescapeString(rss.Channel.Title)
	rss.Channel.Description = html.UnescapeString(rss.Channel.Description)
	for i := range rss.Channel.Items {
		rss.Channel.Items[i].Title = html.UnescapeString(rss.Channel.Items[i].Title)
		rss.Channel.Items[i].Description = html.UnescapeString(rss.Channel.Items[i].Description)
	}
	return &rss, nil
}




//
// Main
//
func main() {

	cfg, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}


	// Adding Commands
	commands := commands{Map: make(map[string]func(s *state, cmd command) error)}
	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)
	commands.register("reset", handlerReset)
	commands.register("users", handlerUsers)
	commands.register("agg", handlerAgg)
	commands.register("help", handlerHelp)
	commands.register("feeds", handlerFeeds)
	commands.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	commands.register("follow", middlewareLoggedIn(handlerFollow))
	commands.register("following", middlewareLoggedIn(handlerFollowing))
	commands.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	commands.register("browse", middlewareLoggedIn(handlerBrowse))
	state := state{config: cfg, cmds: commands}


	// 
	// DB Connection
	//
	db, err := sql.Open("postgres", *cfg.DBUrl)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	state.db = database.New(db)


	//
	// Parse and Run Arguments
	//
	if len(os.Args) < 2 {
		fmt.Println("argument was not given")		
		os.Exit(1)
	}

	command := command{
		Name: os.Args[1],
		Args: os.Args[2:],
	}

	err = commands.run(&state, command)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
