package rdb

type RDB struct {
	Dir string
	DBfilename string
}

func NewRDB(dir, dbfilename string) *RDB {
	return &RDB{
		Dir: dir,
		DBfilename: dbfilename,
	}
}
