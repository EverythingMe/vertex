package main

import (
	"gitlab.doit9.com/backend/vertex"
	_ "gitlab.doit9.com/backend/vertex/vertex-server/example"
)

func init() {

}
func main() {
	vertex.ReadConfigs()
	srv := vertex.NewServer(":9947")
	srv.InitAPIs()
	if err := srv.Run(); err != nil {
		panic(err)
	}

}
