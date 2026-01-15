package logger

import "hash/fnv"

// Module color pool (excluding red, yellow, green, cyan for log levels)
var moduleColors = []string{
	"\033[35m", // Magenta
	"\033[34m", // Blue
	"\033[95m", // Bright magenta
	"\033[94m", // Bright blue
	"\033[96m", // Bright cyan
	"\033[93m", // Bright yellow
	"\033[97m", // Bright white
	"\033[91m", // Bright red
}

// getModuleColor gets color based on module name hash
func getModuleColor(module string) string {
	h := fnv.New32a()
	h.Write([]byte(module))
	return moduleColors[h.Sum32()%uint32(len(moduleColors))]
}
