package types

func (p Plan) IsSettled() bool {
	return p.SettledDenom != ""
}
