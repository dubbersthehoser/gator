package main

import (
	
	"log"
	"fmt"
	"os"
	
	"github.com/dubbersthehoser/gator/internal/config"

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

func main() {


	cfg, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}
	commands := commands{Map: make(map[string]func(s *state, cmd command) error)}
	commands.register("login", handlerLogin)
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

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}



}
