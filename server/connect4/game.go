package connect4

import (
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"time"
)

func NewGame(p1, p2, channelID string) Game {
	rand.Seed(time.Now().UnixNano())
	return &game{
		Board:        newBoard(),
		Player1:      p1,
		Player2:      p2,
		LastMovement: 0,
		Result:       OutcomeNoOutcome,
		Turn:         rand.Intn(2) + 1,
		ChannelID:    channelID,
	}
}

func (g *game) SetPostID(pID string) {
	g.PostID = pID
}

func (g *game) Outcome() int {
	return g.Result
}

func (g *game) GetTurnPlayer() string {
	if g.Turn == Player1 {
		return g.Player1
	}

	return g.Player2
}

func (g *game) Move(movement int) error {
	err := g.Board.Move(movement, g.Turn)
	if err != nil {
		return err
	}

	g.Turn = (g.Turn % 2) + 1
	g.LastMovement = movement

	if g.Board.HasFinished() {
		if g.Board.checkDraw() {
			g.Result = OutcomeDraw
		}
		wonUser := g.Board.HasWon()
		if wonUser != 0 {
			g.Result = wonUser
		}
	}

	return nil
}

func (g *game) Resign(player int) error {
	switch player {
	case Player1:
		g.Result = OutcomePlayer1Resign
	case Player2:
		g.Result = OutcomePlayer2Resign
	default:
		return errors.New("player not playing this game")
	}

	return nil
}

func (g *game) ToJSON() []byte {
	b, _ := json.Marshal(g)
	return b
}

func (g *game) EncodeBoard(w io.Writer) {
	EncodeBoard(w, g.Board, g.LastMovement)
}

func GameFromJSON(b []byte) (Game, error) {
	g := &game{}
	err := json.Unmarshal(b, &g)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (g *game) GetMetadata() (string, string, string, string) {
	return g.ChannelID, g.PostID, g.Player1, g.Player2
}

func (g *game) ValidMovements() []int {
	return g.Board.GetValidMovements()
}
