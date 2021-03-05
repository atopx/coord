package coord

import (
	"errors"
	"math"
	"strconv"
	"strings"
)

// 坐标系
type Coord struct {
	System    string  `json:"system"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

const (
	XPi         = 52.35987755982988      // Pi*3000/180
	Axis        = 6378245.0              // 长半轴
	Offset      = 0.00669342162296594323 // 偏心率平方
	BDLonOffset = 0.0065                 // 百度坐标系经度偏移常量
	BDLatOffset = 0.0060                 // 百度坐标系纬度偏移常量
)

func (c *Coord) Transformation() (coords []Coord) {
	var longitude, latitude float64
	coords = append(coords, Coord{c.System, c.Longitude, c.Latitude})
	switch c.System {
	case "GCJ02":
		longitude, latitude = GCJ02toBD09(c.Longitude, c.Latitude)
		coords = append(coords, Coord{"BD09", longitude, latitude})
		longitude, latitude = GCJ02toWGS84(c.Longitude, c.Latitude)
		coords = append(coords, Coord{"WGS84", longitude, latitude})
	case "BD09":
		longitude, latitude = BD09toGCJ02(c.Longitude, c.Latitude)
		coords = append(coords, Coord{"GCJ02", longitude, latitude})
		longitude, latitude = GCJ02toWGS84(longitude, latitude)
		coords = append(coords, Coord{"WGS84", longitude, latitude})
	case "WGS84":
		longitude, latitude = WGS84toGCJ02(c.Longitude, c.Latitude)
		coords = append(coords, Coord{"GCJ02", longitude, latitude})
		longitude, latitude = GCJ02toBD09(longitude, latitude)
		coords = append(coords, Coord{"BD09", longitude, latitude})
	}
	return coords
}


// 计算偏移量
func delta(lon, lat float64) (float64, float64) {
	dlat, dlon := transform(lon-105.0, lat-35.0)
	radlat := lat / 180.0 * math.Pi
	magic := math.Sin(radlat)
	magic = 1 - Offset*magic*magic
	sqrtmagic := math.Sqrt(magic)
	dlat = (dlat * 180.0) / ((Axis * (1 - Offset)) / (magic * sqrtmagic) * math.Pi)
	dlon = (dlon * 180.0) / (Axis / sqrtmagic * math.Cos(radlat) * math.Pi)
	return lon + dlon, lat + dlat
}

// 转换逻辑
func transform(lon, lat float64) (x, y float64) {
	var lonlat = lon * lat
	var absX = math.Sqrt(math.Abs(lon))
	var lonPi, latPi = lon * math.Pi, lat * math.Pi
	var d = 20.0*math.Sin(6.0*lonPi) + 20.0*math.Sin(2.0*lonPi)
	x, y = d, d
	x += 20.0*math.Sin(latPi) + 40.0*math.Sin(latPi/3.0)
	y += 20.0*math.Sin(lonPi) + 40.0*math.Sin(lonPi/3.0)
	x += 160.0*math.Sin(latPi/12.0) + 320*math.Sin(latPi/30.0)
	y += 150.0*math.Sin(lonPi/12.0) + 300.0*math.Sin(lonPi/30.0)
	x *= 2.0 / 3.0
	y *= 2.0 / 3.0
	x += 2.0*lon + 3.0*lat + 0.2*lat*lat + 0.1*lonlat + 0.2*absX - 100.0
	y += lon + 2.0*lat + 0.1*lon*lon + 0.1*lonlat + 0.1*absX + 300.0
	return x, y
}

// 中国范围：lon: 73.66~135.05  lat: 3.86~53.55
func InChina(lon, lat float64) bool {
	return (73.66 < lon && lon < 135.05) && (3.86 < lat && lat < 53.55)
}

//WGS84toGCJ02 WGS84坐标系->火星坐标系
func WGS84toGCJ02(lon, lat float64) (float64, float64) {
	if !InChina(lon, lat) {
		return lon, lat
	}
	return delta(lon, lat)
}

//GCJ02toBD09 火星坐标系->百度坐标系
func GCJ02toBD09(lon, lat float64) (float64, float64) {
	z := math.Sqrt(lon*lon+lat*lat) + 0.00002*math.Sin(lat*XPi)
	theta := math.Atan2(lat, lon) + 0.000003*math.Cos(lon*XPi)
	return z*math.Cos(theta) + BDLonOffset, z*math.Sin(theta) + BDLatOffset
}

//BD09toGCJ02 百度坐标系->火星坐标系
func BD09toGCJ02(lon, lat float64) (float64, float64) {
	var x = lon - BDLonOffset
	var y = lat - BDLatOffset
	z := math.Sqrt(x*x+y*y) - 0.00002*math.Sin(y*XPi)
	theta := math.Atan2(y, x) - 0.000003*math.Cos(x*XPi)
	return z * math.Cos(theta), z * math.Sin(theta)
}

//GCJ02toWGS84 火星坐标系->WGS84坐标系(微调)
func GCJ02toWGS84(lon, lat float64) (float64, float64) {
	threshold := 0.0000000001 // 设置精准率阙值
	mlon := lon - 0.01
	mlat := lat - 0.01
	plon := lon + 0.01
	plat := lat + 0.01
	var dlon, dlat, wgsLat, wgsLon float64
	for i := 0; i < 10000; i++ {
		wgsLat = (mlat + plat) / 2
		wgsLon = (mlon + plon) / 2
		tmpLon, tmpLat := delta(wgsLon, wgsLat)
		dlon = tmpLon - lon
		dlat = tmpLat - lat
		if math.Abs(dlat) < threshold && math.Abs(dlon) < threshold {
			break
		}
		if dlat > 0 {
			plat = wgsLat
		} else {
			mlat = wgsLat
		}

		if dlon > 0 {
			plon = wgsLon
		} else {
			mlon = wgsLon
		}
	}
	return wgsLon, wgsLat
}

// 分割后的字符串转浮点数经纬度: ("115.668055","34.449162") => (115.668055,34.449162)
func DoubleStringToFloat64Coord(Lon, Lat string) (longitude, latitude float64, err error) {
	longitude, err = strconv.ParseFloat(strings.TrimSpace(Lon), 64)
	if err == nil {
		latitude, err = strconv.ParseFloat(strings.TrimSpace(Lat), 64)
	}
	return longitude, latitude, err
}

// location字符串转浮点数经纬度: "115.668055,34.449162" => (115.668055,34.449162)
func LocationToFloat64Coord(location string) (longitude, latitude float64, err error) {
	var message = "不是有效的Location字符串"
	records := strings.Split(location, ",")
	if len(records) < 2 {
		return longitude, latitude, errors.New(message)
	}
	return DoubleStringToFloat64Coord(records[0], records[1])
}
