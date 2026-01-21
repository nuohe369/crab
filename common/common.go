package common

import (
	"github.com/nuohe369/crab/pkg/logger"
)

var log = logger.NewSystem("common")

// Init initializes the common business layer
// Init 初始化通用业务层
func Init() {
	log.Info("Initializing common business layer...")
}
