package common

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
