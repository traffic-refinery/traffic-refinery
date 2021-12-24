package stats

type Stats interface {
	Type() string
	Init() error
	Run() []byte
}
