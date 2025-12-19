package serilizers

type Serilizer interface {
	Serilize(config any) error
}
