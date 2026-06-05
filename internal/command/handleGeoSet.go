package handler

import (
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
	"strconv"
	"fmt"
)

var MaxLongitude = 180.0
var MinLongitude = -180.0
var MaxLatitude = 85.05112878
var MinLatitude = -85.05112878


func _interleaveits(x, y uint32) uint64 {
	var B = [5]uint64{
		0x5555555555555555,
		0x3333333333333333,
		0x0F0F0F0F0F0F0F0F, 
		0x00FF00FF00FF00FF,
		0x0000FFFF0000FFFF,
	}

	var S = [5]uint8{1, 2, 4, 8, 16,}

	var x64 = uint64(x)
	var y64 = uint64(y)

	for i := 4; i >= 0; i-- {
		x64 = (x64 | (x64 << S[i])) & B[i]
		y64 = (y64 | (y64 << S[i])) & B[i]
	}
	return x64 | (y64 << 1)
}

func _geoHashEncode(longitude, latitude float64) uint64 {

	lat_offset := (latitude - MinLatitude) / (MaxLatitude - MinLatitude)

	lon_offset := (longitude - MinLongitude) / (MaxLongitude - MinLongitude)

	lat_offset *= (1 << 26)
	lon_offset *= (1 << 26)

	return _interleaveits(uint32(lon_offset), uint32(lat_offset))
}



func handleGEOADD(st *store.ExpireMap, args []string) []byte {
	if len(args) != 5 {
		return resp.BuildError("ERR wrong number of arguments for 'GEOADD' command")
	}

	key := args[1]
	longitude, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return resp.BuildError("ERR value is not a valid float")
	}
	latitude, err := strconv.ParseFloat(args[3], 64)
	if err != nil {
		return resp.BuildError("ERR value is not a valid float")
	}

	if longitude < MinLongitude || longitude > MaxLongitude {
		return resp.BuildError(fmt.Sprintf("ERR invalid longitude,latitude pair %.6f, %.6f", longitude, latitude))
	}
	if latitude < MinLatitude || latitude > MaxLatitude {
		return resp.BuildError(fmt.Sprintf("ERR invalid longitude,latitude pair %.6f, %.6f", longitude, latitude))
	}
	member := args[4]

	geoHash := _geoHashEncode(longitude, latitude)
	val, err := st.ZAdd(key, float64(geoHash), member)

	if err != nil {
		return resp.BuildError("ERR could not add geo data")
	}
	return resp.BuildInteger(val)
}
