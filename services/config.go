package services

import "kasen/config"

func GetServiceConfig() config.Service {
	return config.GetService()
}

func UpdateMeta(v *config.Meta) error {
	config.SetMeta(*v)
	return config.Save()
}

func UpdateServiceConfig(v *config.Service) error {
	config.SetService(*v)
	return config.Save()
}
