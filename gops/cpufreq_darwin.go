//go:build darwin

package gops

func readCPUStaticLive() cpuStatic   { return cpuStatic{} }
func readCPUDynamicLive() cpuDynamic { return cpuDynamic{} }
