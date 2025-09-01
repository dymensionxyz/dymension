package types

// GraduationStatus represents the computed state of a plan
type GraduationStatus int32

const (
	GraduationStatus_PRE_GRADUATION GraduationStatus = 0
	GraduationStatus_POOL_CREATED   GraduationStatus = 1
	GraduationStatus_SETTLED        GraduationStatus = 2
)

// graduitstatus st4ring
var GraduationStatusString = map[GraduationStatus]string{
	GraduationStatus_PRE_GRADUATION: "pre_graduation",
	GraduationStatus_POOL_CREATED:   "pool_created",
	GraduationStatus_SETTLED:        "settled",
}

// GetGraduationStatusString returns the string representation of the graduation status
func (s GraduationStatus) String() string {
	return GraduationStatusString[s]
}

// GetGraduationStatus computes the current status from field presence
func (p Plan) GetGraduationStatus() GraduationStatus {
	if p.GraduatedPoolId == 0 {
		return GraduationStatus_PRE_GRADUATION
	}
	if p.SettledDenom == "" {
		return GraduationStatus_POOL_CREATED
	}
	return GraduationStatus_SETTLED
}

// Helper methods using computed status
func (p Plan) PreGraduation() bool {
	return p.GetGraduationStatus() == GraduationStatus_PRE_GRADUATION
}

func (p Plan) IsGraduated() bool {
	return p.GetGraduationStatus() == GraduationStatus_POOL_CREATED
}

func (p Plan) IsSettled() bool {
	return p.GetGraduationStatus() == GraduationStatus_SETTLED
}
