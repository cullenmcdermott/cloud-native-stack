package collectors

type Collector interface {
	Collect(config any) ([]Configuraion, error)
}

type Configuraion struct {
	Type string
	Data any
}
