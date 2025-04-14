package common

func IsInThailand(lat, long float64) bool {
	// Thailand's approximate latitude and longitude bounds
	return lat >= 5.6 && lat <= 20.5 && long >= 97.3 && long <= 105.7
}
