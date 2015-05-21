package main

import (
	"gitlab.doit9.com/backend/vertex"
	_ "gitlab.doit9.com/backend/vertex/vertex-server/example"
)

func main() {

	srv := vertex.NewServer(":9947")

	if err := srv.Run(); err != nil {
		panic(err)
	}

}
