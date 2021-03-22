package service

func init() {
	// registry
	types = make(map[string]sType)
	services = make(map[int64]Service)

}
