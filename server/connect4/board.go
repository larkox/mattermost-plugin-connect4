package connect4

import "errors"

func newBoard() board {
	b := board{}
	for i := 0; i < columns; i++ {
		column := []int{}
		for j := 0; j < rows; j++ {
			column = append(column, 0)
		}
		b = append(b, column)
	}
	return b
}

func (b board) HasFinished() bool {
	return b.checkDraw() || b.HasWon() != 0
}

func (b board) HasWon() int {
	for i, column := range b {
		for j, square := range column {
			if square != 0 {
				win := b.checkWin(i, j, square)
				if win {
					return square
				}
			}
		}
	}
	return 0
}

func (b board) checkWin(i, j, player int) bool {
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			for d := 1; d <= 3; d++ {
				if dx == 0 && dy == 0 {
					break
				}
				x := i + dx*d
				y := j + dy*d
				if x < 0 || x >= columns || y < 0 || y >= rows {
					break
				}
				if b[x][y] != player {
					break
				}
				if d == 3 {
					return true
				}
			}
		}
	}
	return false
}

func (b board) checkDraw() bool {
	for _, column := range b {
		if column[0] == 0 {
			return false
		}
	}
	return true
}

func (b board) Move(column, player int) error {
	columnIndex := column - 1
	if columnIndex > columns || columnIndex < 0 {
		return errors.New("Not such column")
	}

	if b[columnIndex][0] != 0 {
		return errors.New("Column full")
	}

	toUpdate := len(b[columnIndex]) - 1
	for i, square := range b[columnIndex] {
		if square != 0 {
			toUpdate = i - 1
			break
		}
	}
	b[columnIndex][toUpdate] = player
	return nil
}
