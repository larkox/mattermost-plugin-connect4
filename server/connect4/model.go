package connect4

import "io"

type board [][]int

type Game interface {
	SetPostID(pID string)
	Outcome() int
	GetTurnPlayer() string
	Move(movement int) error
	Resign(player int) error
	ToJSON() []byte
	GetMetadata() (string, string, string, string)
	EncodeBoard(w io.Writer)
	ValidMovements() []int
}

type game struct {
	Board        board
	Player1      string
	Player2      string
	LastMovement int
	Turn         int
	ChannelID    string
	PostID       string
	Result       int
}
