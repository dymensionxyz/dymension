package types

func (h *HookEIBCtoHL) ValidateBasic() error {
	return nil
}

func (h *HookHLtoIBC) ValidateBasic() error {
	err := h.Transfer.ValidateBasic()
	if err != nil {
		return err
	}
	// TODO: can check timeout height is zero(?)

	return nil
}
