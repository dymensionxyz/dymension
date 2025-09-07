package types

// GraduationStatus represents the computed state of a plan
type GraduationStatus int32

const (
	GraduationStatus_PRE_GRADUATION GraduationStatus = 0
	GraduationStatus_GRADUATED      GraduationStatus = 1
	GraduationStatus_SETTLED        GraduationStatus = 2
)

// graduation status string
var GraduationStatusString = map[GraduationStatus]string{
	GraduationStatus_PRE_GRADUATION: "pre_graduation",
	GraduationStatus_GRADUATED:      "graduated",
	GraduationStatus_SETTLED:        "settled",
}

// GetGraduationStatusString returns the string representation of the graduation status
func (s GraduationStatus) String() string {
	return GraduationStatusString[s]
}

// GetGraduationStatus computes the current status from field presence
func (p Plan) GetGraduationStatus() GraduationStatus {
	if p.SettledDenom != "" {
		return GraduationStatus_SETTLED
	}
	if p.GraduatedPoolId != 0 {
		return GraduationStatus_GRADUATED
	}
	return GraduationStatus_PRE_GRADUATION
}

// Helper methods using computed status
func (p Plan) PreGraduation() bool {
	return p.GetGraduationStatus() == GraduationStatus_PRE_GRADUATION
}

func (p Plan) IsGraduated() bool {
	return p.GetGraduationStatus() == GraduationStatus_GRADUATED
}

func (p Plan) IsSettled() bool {
	return p.GetGraduationStatus() == GraduationStatus_SETTLED
}
