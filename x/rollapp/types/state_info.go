package types

func (m *StateInfo) BlockDescriptorByHeight(height uint64) BlockDescriptor {
	return m.BDs.BD[height-m.StartHeight]
}
