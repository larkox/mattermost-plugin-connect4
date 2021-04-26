package connect4

func (b board) GetValidMovements() []int {
	out := []int{}
	for i, column := range b {
		if column[0] == 0 {
			out = append(out, i+1)
		}
	}
	return out
}
