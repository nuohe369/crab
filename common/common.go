package common

import (
	"github.com/nuohe369/crab/common/service"
	"github.com/nuohe369/crab/pkg/logger"
)

var log = logger.NewSystem("common")

// Init initializes the common business layer.
func Init() {
	log.Info("Initializing common business layer...")

	// Initialize WebSocket service
	service.InitWS()
}
