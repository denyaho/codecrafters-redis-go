package rdb

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

type RDBParser struct {
	data []byte
	pos  int
}

func NewRDBParser(data []byte) *RDBParser {
	return &RDBParser{
		data: data,
		pos:  0,
	}
}

func (p *RDBParser) readHeader(data []byte, r *RDB) error {
	if len(data) < 9 {
		return fmt.Errorf("invalid RDB file: too short to contain header")
	}

	magic := string(data[0:5])
	if magic != "REDIS" {
		return fmt.Errorf("invalid RDB file: missing REDIS header")
	}

	versionstr := string(data[5:9])
	version, err := strconv.Atoi(versionstr)
	if err != nil {
		return fmt.Errorf("invalid RDB file: invalid version number")
	}
	r.version = version
	p.pos += 9

	return nil
}

func (p *RDBParser) readLength() (int, error) {
	top := p.data[p.pos]
	p.pos++
	fmt.Printf("Read byte : %v\n", top>>6)
	switch top >> 6 {
		case 0b00:
			return int(top & 0b00111111), nil
		case 0b01:
			next := p.data[p.pos]
			p.pos++
			return int(top & 0b00111111) << 8 | int(next), nil
		case 0b10:
			if p.pos + 4 > len(p.data) {
				return 0, fmt.Errorf("invalid RDB file: unexpected end of data while reading length")
			}
			next := binary.BigEndian.Uint32(p.data[p.pos : p.pos+4])
			p.pos += 4
			return int(next), nil			
	}
	return 0, fmt.Errorf("invalid RDB file: invalid length encoding")
}

func (p *RDBParser) readEncodedString() (string, error) {
	length, err := p.readLength()
	if err != nil {
		fmt.Printf("Failed to read length of encoded string: %v\n", err)
		return "", err
	}
	if p.pos + length > len(p.data) {
		return "", fmt.Errorf("invalid RDB file: unexpected end of data while reading string")
	}
	str := string(p.data[p.pos : p.pos+length])
	p.pos += length
	return str, nil
}

func (p *RDBParser) readMetaData(r *RDB) error {
	metadata := make(map[string]string)
	p.pos++
	key, err := p.readEncodedString()
	if err != nil {
		fmt.Println("Failed to read metadata key: ", err.Error())
		return err
	}
	value, err := p.readEncodedString()
	if err != nil {
		fmt.Println("Failed to read metadata value: ", err.Error())
		return err
	}
	metadata[key] = value
	r.metadata = metadata
	return nil
}

func (p *RDBParser) readByte(n int) []byte{ 
	b := p.data[p.pos : p.pos+n]
	p.pos += n
	return b
}

func (p *RDBParser) readDBSelect(r *RDB) error {
	p.pos++
	dbIndex, err := p.readLength()
	if err != nil {
		return err
	}
	r.DBs = append(r.DBs, DBInfo{DBIndex: dbIndex})
	return nil
}

func (p *RDBParser) readResizeDB(r *RDB) error {
	p.pos++
	dbSize, err := p.readLength()
	if err != nil {
		return err
	}
	expireSize, err := p.readLength()
	if err != nil {
		return err
	}
	if len(r.DBs) > 0 {
		r.DBs[len(r.DBs)-1].HashTableSize = dbSize
		r.DBs[len(r.DBs)-1].ExpireTableSize = expireSize
	}
	return nil
}

func (p *RDBParser) handleKeyValuePair(st *store.ExpireMap, expireAt time.Duration) error {
	var key, value string
	var err error
	switch p.data[p.pos] {
		case 0x00:
			p.pos++
			key, err = p.readEncodedString()
			if err != nil {
				return err
			}
			value, err = p.readEncodedString()
			if err != nil {
				return err
			}
			st.Set(key, value, expireAt)
	}
	fmt.Printf("Key-value pair loaded: key=%s, value=%s, expireAt=%v\n", key, value, expireAt)
	return nil
}

func (p *RDBParser) readExpiryms(r *RDB, st *store.ExpireMap) error {
	p.pos++
	expiryms := binary.LittleEndian.Uint64(p.data[p.pos : p.pos+8])
	p.pos += 8
	expireAt := time.Until(time.UnixMilli(int64(expiryms)))
	return p.handleKeyValuePair(st, expireAt)
}

func (p *RDBParser) readExpirysec(r *RDB, st *store.ExpireMap) error {
	p.pos++
	expirysec := binary.LittleEndian.Uint32(p.data[p.pos : p.pos+4])
	p.pos += 4
	expireAt := time.Until(time.Unix(int64(expirysec), 0))
	return p.handleKeyValuePair(st, expireAt)
}

func (p *RDBParser) Parse(r *RDB, st *store.ExpireMap) error {
	err := p.readHeader(p.data, r)
	if err != nil {
		return err
	}
	fmt.Printf("RDB version: %d\n", r.version)
	for p.pos < len(p.data) {
		fmt.Printf("Parsing RDB at position %d\n", p.pos)
		switch p.data[p.pos] {
			case 0xFA:
				fmt.Println("Reading 0xFA metadata")
				err := p.readMetaData(r)
				if err != nil {
					return err
				}
				fmt.Printf("RDB metadata: %v\n", r.metadata["redis-ver"])
			case 0xFB:
				fmt.Println("Reading 0xFB resize DB")
				err := p.readResizeDB(r)
				if err != nil {
					return err
				}
			case 0xFC:
				fmt.Println("Reading 0xFC expiry ms")
				err := p.readExpiryms(r, st)
				if err != nil {
					return err
				}
			case 0xFD:
				fmt.Println("Reading 0xFD expiry sec")
				err := p.readExpirysec(r, st)
				if err != nil {
					return err
				}
			case 0xFE:
				fmt.Println("Reading 0xFE DB select")
				err := p.readDBSelect(r)
				if err != nil {
					return err
				}
			case 0xFF:
				p.pos++
				r.checksum = p.readByte(8)
				return nil
			default:
				if p.data[p.pos] == 0x00 {
					err := p.handleKeyValuePair(st, 0)
					if err != nil {
						return err
					}
				} else {
					return fmt.Errorf("invalid RDB file: unexpected byte %x at position %d", p.data[p.pos], p.pos)
				}
		}
	}
	return nil

}
