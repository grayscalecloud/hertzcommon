package kvconfig

import "os"

func InitConfEnvs() {
	port := os.Getenv("PORT")
	if len(port) == 0 {

	}
	kvKey := os.Getenv("KV_KEY")
	if len(kvKey) == 0 {

	}
	_ = os.Getenv("REGISTRY_ADDRESS")

	_ = os.Getenv("REGISTRY_ADDRESS_USERNAME")

	_ = os.Getenv("REGISTRY_ADDRESS_PASSWORD")

}
