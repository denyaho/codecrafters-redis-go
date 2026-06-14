package aof

type AOF struct {
	Dir            string
	AppendOnly     string
	AppendDirname  string
	AppendFilename string
	AppendFsync    string
}

func NewAOF(dir string) *AOF {
	return &AOF{
		Dir: dir,
		AppendOnly: "no",
		AppendDirname: "appendonlydir",
		AppendFilename: "appendonly.aof",
		AppendFsync: "everysec",
	}
}