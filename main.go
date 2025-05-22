package main

import (
"sync"

"github.com/PretendoNetwork/super-mario-maker/nex"
"github.com/PretendoNetwork/super-mario-maker/nex/game-mode-manager"
)

var wg sync.WaitGroup

func main() {
// Initialize the GameModeManager
game_mode_manager.Init()

wg.Add(2)

// TODO - Add gRPC server
go nex.StartAuthenticationServer()
go nex.StartSecureServer()

wg.Wait()
}
