package cachex

type Cacher interface {
	Get(string) (any, bool)
	Set(string, any)
	Del(...string) int
}
