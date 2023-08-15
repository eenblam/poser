package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
)

var ErrInvalidState = errors.New("invalid game state")

// This can be made a room config variable later
const maxRounds = 2

// Player states
type PlayerState struct {
	// Turns taken by player so far
	TurnsTaken int
	// Index of player's number in Game.Players
	Index int
}

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
	// Current State of the game (Waiting, GettingPrompt, Drawing, Voting, PoserGuessing)
	State State
	// Array of player numbers
	Players []int
	// Map of playerNumber -> PlayerState
	PlayerStates map[int]*PlayerState
	// Current round
	Round int
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
	g.PlayerStates = nil
	g.Muse = 0
	g.Poser = 0
	g.Drawing = 0
	g.Round = 0
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

	g.PlayerStates = make(map[int]*PlayerState)
	for i, p := range players {
		g.PlayerStates[p] = &PlayerState{
			TurnsTaken: 0,
			Index:      i,
		}
	}

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

func (g *Game) EndTurn(player int) error {
	if g.State != Drawing {
		return ErrInvalidState
	}

	if player != g.Drawing {
		return fmt.Errorf("player %d is not drawing", player)
	}

	// Increment turn count for player
	g.PlayerStates[player].TurnsTaken++

	// Find next player
	nextIndex := (g.PlayerStates[player].Index + 1) % len(g.Players)
	nextPlayer, ok := g.PlayerStates[g.Players[nextIndex]]
	if !ok {
		log.Println("index of next player not found in PlayerStates")
		return fmt.Errorf("player #%d is next, but not found in PlayerStates", g.Players[nextIndex])
	}
	if nextPlayer.TurnsTaken == maxRounds {
		// We've wrapped back around, everyone has played.
		g.State = Voting
		return nil
	}
	// Otherwise, advance to next player
	g.Drawing = g.Players[nextIndex]
	return nil
}
