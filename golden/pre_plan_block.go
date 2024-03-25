package golden

type PrePlanBlock interface {
	ExecuteBeforePlan() error
}
