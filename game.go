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

// Value for PlayerState.Vote when player has not voted
const NoVote = -1

// Player states
type PlayerState struct {
	// Turns taken by player so far
	TurnsTaken int
	// Index of player's number in Game.Players
	Index int
	// Vote is the PlayerNumber of the player this player voted for
	Vote int
	// VotesAgainst is the number of players who think this player is the Poser
	VotesAgainst int
}

// Game states
type State string

const (
	Waiting       State = "Waiting"
	GettingPrompt State = "GettingPrompt"
	Drawing       State = "Drawing"
	Voting        State = "Voting"
	PoserGuessing State = "PoserGuessing"
	PoserWon      State = "PoserWon"
	PoserWonByTie State = "PoserWonByTie"
	PoserLost     State = "PoserLost"
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
			// Set Vote to roomSize, which is 1 + max player number
			Vote:         NoVote,
			VotesAgainst: 0,
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

func (g *Game) Vote(playerNumber int, voteNumber int) error {
	if g.State != Voting {
		return ErrInvalidState
	}
	if g.PlayerStates[playerNumber].Vote != NoVote {
		// Player already voted
		return ErrVotedTwice
	}
	for _, p := range g.Players {
		if p != voteNumber {
			continue
		}
		// Valid vote! Set it.
		g.PlayerStates[playerNumber].Vote = voteNumber
		g.PlayerStates[voteNumber].VotesAgainst++

		for _, ps := range g.PlayerStates {
			if ps.Vote == NoVote { // Someone hasn't voted, continue
				return nil
			}
		}
		{ // Everyone has voted! Tally them up and find the winner.
			// Not trying to be hyper-efficient here, N is tiny.
			maxVote := 0
			for _, ps := range g.PlayerStates {
				if ps.VotesAgainst > maxVote {
					maxVote = ps.VotesAgainst
				}
			}
			// Find all players with max votes to detect a tie
			tiedPlayers := make([]int, 0)
			for pn, ps := range g.PlayerStates {
				if ps.VotesAgainst == maxVote {
					tiedPlayers = append(tiedPlayers, pn)
				}
			}
			if len(tiedPlayers) > 1 { // Tie
				//TODO tie fails
				g.State = PoserWonByTie
			} else if tiedPlayers[0] == g.Poser { // Caught
				g.State = PoserGuessing
			} else { // Not caught
				g.State = PoserWon
			}
			return nil
		}
	}
	return ErrInvalidVote
}
