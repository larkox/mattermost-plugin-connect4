package connect4

import (
	"io"
	"strconv"

	svg "github.com/ajstarks/svgo"
)

const (
	rows        = 6
	columns     = 7
	sqrSize     = 80
	boardWidth  = sqrSize * columns
	boardHeight = sqrSize * (rows + 1)
	margin      = sqrSize / 10
	radio       = (sqrSize / 2) - margin

	textSize = sqrSize - margin

	boardColor   = "#1161ea"
	player1Color = "#ffe042"
	player2Color = "#fe1614"
	emptyColor   = "#ffffff"
	textColor    = "#000000"
)

func EncodeBoard(w io.Writer, board [][]int, lastMovement int) {
	canvas := svg.New(w)
	canvas.Start(boardWidth, boardHeight)
	canvas.Rect(0, 0, boardWidth, boardHeight, "fill: "+boardColor)

	for i := 0; i < columns; i++ {
		x := i * sqrSize
		highlighted := false
		for j := 0; j < rows; j++ {
			y := j * sqrSize
			style := "fill: "
			player := board[i][j]
			switch player {
			case 1:
				color := player1Color
				style += color
			case 2:
				color := player2Color
				style += color
			default:
				style += emptyColor
			}
			canvas.Circle(x+sqrSize/2, y+sqrSize/2, radio, style)
			if i == lastMovement-1 && !highlighted && player != 0 {
				canvas.Circle(x+sqrSize/2, y+sqrSize/2, radio/2, "fill: #000000")
				highlighted = true
			}
		}
		textStyle := "dominant-baseline:middle;text-anchor:middle;fill:" + textColor + ";font-size: " + strconv.Itoa(textSize) + "px"
		canvas.Text(x+sqrSize/2, (rows)*sqrSize+sqrSize/2, strconv.Itoa(i+1), textStyle)
	}
	canvas.End()
}
