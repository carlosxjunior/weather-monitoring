package cities

type City struct {
	Name      string
	Latitude  float64
	Longitude float64
}

var All = []City{
	{Name: "Campinas", Latitude: -22.9056, Longitude: -47.0608},
	{Name: "Florianópolis", Latitude: -27.5954, Longitude: -48.5480},
	{Name: "Rio de Janeiro", Latitude: -22.9068, Longitude: -43.1729},
	{Name: "Brasília", Latitude: -15.7801, Longitude: -47.9292},
	{Name: "New York", Latitude: 40.7128, Longitude: -74.0060},
	{Name: "Berlin", Latitude: 52.5200, Longitude: 13.4050},
	{Name: "Madrid", Latitude: 40.4168, Longitude: -3.7038},
	{Name: "Tokyo", Latitude: 35.6762, Longitude: 139.6503},
	{Name: "Shanghai", Latitude: 31.2304, Longitude: 121.4737},
	{Name: "Melbourne", Latitude: -37.8136, Longitude: 144.9631},
}
