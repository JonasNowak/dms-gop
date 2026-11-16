//go:build darwin

package gops

type platformSources struct{}

func defaultPlatformSources() platformSources { return platformSources{} }

func (self *GopsUtil) closePlatform() {}
