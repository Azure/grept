package pkg

type planAction interface {
	ExecuteDuringPlan() error
}
