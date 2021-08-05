package telemetrybuffer

type Item struct {
	Component string
	Measure   string
	Value     map[string]interface{}
	Labels    map[string]string
}
