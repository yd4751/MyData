package utils

import (
	"crypto/md5"
	"encoding/hex"
	"math"
	"math/rand"
	"time"
)

func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func RandInt(min, max int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return min + r.Intn(max-min+1)
}

func RandFloat(min, max float64) float64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return min + r.Float64()*(max-min)
}

func Clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func Distance(x1, y1, z1, x2, y2, z2 float64) float64 {
	return math.Sqrt(math.Pow(x1-x2, 2) + math.Pow(y1-y2, 2) + math.Pow(z1-z2, 2))
}

func Distance2D(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt(math.Pow(x1-x2, 2) + math.Pow(y1-y2, 2))
}

func Abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func Contains(arr []int64, target int64) bool {
	for _, v := range arr {
		if v == target {
			return true
		}
	}
	return false
}

func RemoveElement(arr []int64, target int64) []int64 {
	for i, v := range arr {
		if v == target {
			return append(arr[:i], arr[i+1:]...)
		}
	}
	return arr
}

func GetChunkPos(x, y float64, chunkSize float64) (int, int) {
	return int(math.Floor(x / chunkSize)), int(math.Floor(y / chunkSize))
}

func GetChunkCenter(chunkX, chunkY int, chunkSize float64) (float64, float64) {
	return float64(chunkX)*chunkSize + chunkSize/2, float64(chunkY)*chunkSize + chunkSize/2
}
