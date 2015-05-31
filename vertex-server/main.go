package main

import (
	"github.com/dvirsky/go-pylog/logging"
	"gitlab.doit9.com/server/vertex"
	_ "gitlab.doit9.com/server/vertex/vertex-server/example"
)

func init() {

}
func main() {
	vertex.ReadConfigs()

	logging.SetMinimalLevelByName(vertex.Config.Server.LoggingLevel)
	srv := vertex.NewServer(vertex.Config.Server.ListenAddr)
	srv.InitAPIs()
	if err := srv.Run(); err != nil {
		panic(err)
	}

}
