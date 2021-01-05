package libbrctl4go

type BRCtl4Go interface {
	AddInterface(string) error
	DeleteInterface(string) error
}