package mode

import "os"

const (
	EnvMode         = "mode"
	DevelopmentMode = "development"
	ProductionMode  = "production"
	TestMode        = "test"
)

var mode = DevelopmentMode

func init() {
	mode := os.Getenv(EnvMode)
	Set(mode)
}

func Set(value string) {
	if value == "" {
		value = DevelopmentMode
	}

	switch value {
	case DevelopmentMode,
		ProductionMode,
		TestMode:
		mode = value
	default:
		panic("unknown mode: " + value)
	}
}

func Get() string {
	return mode
}
