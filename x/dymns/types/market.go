package types

const (
	// NameOrder is an alias variable of OrderType_OT_DYM_NAME
	NameOrder = OrderType_OT_DYM_NAME

	// AliasOrder is an alias variable of OrderType_OT_ALIAS
	AliasOrder = OrderType_OT_ALIAS
)

func (x OrderType) FriendlyString() string {
	switch x {
	case OrderType_OT_DYM_NAME:
		return "Dym-Name"
	case OrderType_OT_ALIAS:
		return "Alias"
	default:
		return "Unknown"
	}
}
