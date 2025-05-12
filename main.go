package main

import (
	
	"log"
	"fmt"
	"os"
	"database/sql"
	"context"
	"time"
	
	"github.com/dubbersthehoser/gator/internal/config"
	"github.com/dubbersthehoser/gator/internal/database"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type state struct {
	config *config.Config
	db     *database.Queries
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


func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("no username was given for login")
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
	bgcon := context.Background()
	id := uuid.New().String()
	pramUser := database.CreateUserParams{
		ID:        id,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
	}

	_, err := s.db.GetUser(bgcon, name)
	if err == nil {
		return fmt.Errorf("already registered")
	}

	_, err = s.db.CreateUser(bgcon, pramUser)
	if err != nil {
		return err
	}

	s.config.SetUser(name)

	fmt.Printf("%s user was created\n", name)
	fmt.Printf("DEBUG: %v\n", pramUser)


	return nil
}

func main() {

	cfg, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}
	commands := commands{Map: make(map[string]func(s *state, cmd command) error)}
	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)
	state := state{config: cfg}

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

	db, err := sql.Open("postgres", *cfg.DBUrl)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	state.db = database.New(db)
}
