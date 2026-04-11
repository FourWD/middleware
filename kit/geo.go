package kit

import "math"

func DistanceMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusMeters = 6371000.0

	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusMeters * c
}

func IsInThailand(lat, long float64) bool {
	return lat >= 5.6 && lat <= 20.5 && long >= 97.3 && long <= 105.7
}

func IsInThailandBox(lat, lng float64) bool {
	const (
		thMinLat = 5.60
		thMaxLat = 20.55
		thMinLng = 97.30
		thMaxLng = 105.90
	)

	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return false
	}

	return lat >= thMinLat && lat <= thMaxLat && lng >= thMinLng && lng <= thMaxLng
}
