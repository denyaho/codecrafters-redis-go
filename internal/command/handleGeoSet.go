package handler

import (
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
	"strconv"
	"fmt"
)

const (
	MAXLONITUDE = 180.0
	MINLONITUDE = -180.0
	MAXLATITUDE = 85.05112878
	MINLATITUDE = -85.05112878
)

func _decodeGeoHash(geoHash uint64) (float64, float64) {
	var B = [6]uint64{
		0x5555555555555555,
		0x3333333333333333,
		0x0F0F0F0F0F0F0F0F, 
		0x00FF00FF00FF00FF,
		0x0000FFFF0000FFFF,
		0x00000000FFFFFFFF,
	}

	var S = [6]uint8{0, 1, 2, 4, 8, 16,}

	var x64 = geoHash
	var y64 = geoHash >> 1

	x64 = x64 & B[0]
	y64 = y64 & B[0]

	for i := 1; i < 6; i++ {
		x64 = (x64 | (x64 >> S[i])) & B[i]
		y64 = (y64 | (y64 >> S[i])) & B[i]
	}
	
	latitude := float64(x64) / (1 << 26) * (MAXLATITUDE - MINLATITUDE) + MINLATITUDE
	longitude := float64(y64) / (1 << 26) * (MAXLONITUDE - MINLONITUDE) + MINLONITUDE
	
	return longitude, latitude
}



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

	lat_offset := (latitude - MINLATITUDE) / (MAXLATITUDE - MINLATITUDE)

	lon_offset := (longitude - MINLONITUDE) / (MAXLONITUDE - MINLONITUDE)

	lat_offset *= (1 << 26)
	lon_offset *= (1 << 26)

	return _interleaveits(uint32(lat_offset), uint32(lon_offset))
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

	if longitude < MINLONITUDE || longitude > MAXLONITUDE {
		return resp.BuildError(fmt.Sprintf("ERR invalid longitude,latitude pair %.6f, %.6f", longitude, latitude))
	}
	if latitude < MINLATITUDE || latitude > MAXLATITUDE {
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

func handleGEOPOS(st *store.ExpireMap, args []string) []byte {
	if len(args) < 3{
		return resp.BuildError("ERR wrong number of arguments for 'GEOPOS' command")
	}
	key := args[1]

	results := make([][]string, len(args)-2)
	for i := 2; i < len(args); i++ {
		member := args[i]
		val, err := st.ZGet(key, member)
		if err != nil {
			return resp.BuildError("ERR could not get geo data")
		}

		if val == -1 {
			results[i-2] = []string{}
			continue
		}
		longitude, latitude := _decodeGeoHash(uint64(val))
		longitudeStr := strconv.FormatFloat(longitude, 'f', 12, 64)
		latitudeStr := strconv.FormatFloat(latitude, 'f', 12, 64)
		results[i-2] = []string{longitudeStr, latitudeStr}
	}
	return resp.BuildArrayOfArrays(results)
}
