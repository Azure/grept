package fixes

type Fix interface {
	ApplyFix() error
}
