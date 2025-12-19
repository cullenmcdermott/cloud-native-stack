package snapshotter

type Snapshotter interface {
	Run(config any) error
}
