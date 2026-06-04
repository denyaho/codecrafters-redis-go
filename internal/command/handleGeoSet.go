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
	val, err := st.GeoAdd(key, longitude, latitude, member)
	if err != nil {
		return resp.BuildError("ERR could not add geo data")
	}
	return resp.BuildInteger(val)
}
