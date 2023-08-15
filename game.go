package main

import (
	"errors"
	"fmt"
	"math/rand"
)

var ErrInvalidState = errors.New("invalid game state")

// Game states
type State string

const (
	Waiting       State = "Waiting"
	GettingPrompt State = "GettingPrompt"
	Drawing       State = "Drawing"
	Voting        State = "Voting"
	PoserGuessing State = "PoserGuessing"
	// ValidatingGuess?
)

// Player roles
type Game struct {
	State State
	// Array of player numbers
	Players []int
	// Index of Muse
	Muse int
	// Index of Poser
	Poser int
	// Index of player who is currently drawing
	Drawing int
	// Prompt for current game
	Prompt string
}

// Abort game, resetting values to defaults.
func (g *Game) Abort() {
	g.State = Waiting
	g.Players = nil
	g.Muse = 0
	g.Poser = 0
	g.Drawing = 0
}

// IsJoinable checks if game can currently be joined by a user.
func (g *Game) IsJoinable() bool {
	return g.State == Waiting
}

// Game expects a list of player numbers.
func (g *Game) Start(players []int) error {
	if g.State != Waiting {
		return ErrGameInProgress
	}

	if len(players) < 2 {
		return fmt.Errorf("not enough players to start game")
	}

	g.Players = make([]int, len(players))
	copy(g.Players, players)

	// Make a separate copy we can mangle
	choices := make([]int, len(players))
	copy(choices, players)

	// Pick Muse
	m := rand.Intn(len(choices))
	g.Muse = choices[m]
	// Pick first player (can be Muse!)
	g.Drawing = choices[rand.Intn(len(choices))]
	// Drop Muse from choices
	choices = append(choices[:m], choices[m+1:]...)
	// Pick Poser (after removing Muse)
	g.Poser = choices[rand.Intn(len(choices))]

	g.State = GettingPrompt
	return nil
}

func (g *Game) SetPrompt(prompt string) error {
	if g.State != GettingPrompt {
		return ErrInvalidState
	}

	g.Prompt = prompt
	g.State = Drawing
	return nil
}
