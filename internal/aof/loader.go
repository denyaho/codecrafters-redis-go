package aof

type AOF struct {
	Dir            string
	Appendonly     bool
	Appenddirname  string
	Appendfilename string
	Appendfsync    string
}

func NewAOF(dir string) *AOF {
	return &AOF{
		Dir: dir,
		Appendonly: false,
		Appenddirname: "appendonly",
		Appendfilename: "appendonly.aof",
		Appendfsync: "everysec",
	}
}