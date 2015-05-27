package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dvirsky/go-pylog/logging"

	"gitlab.doit9.com/backend/vertex"
	_ "gitlab.doit9.com/backend/vertex/vertex-server/example"
)

func init() {

}
func main() {

	serverAddr := flag.String("server", "http://127.0.0.1:9947", "The server URL to connect to")
	apiName := flag.String("api", "", "The API we want to test")
	category := flag.String("category", "all", "The test category we want to run [all|critical|warning]")
	format := flag.String("format", "text", "Result Output Format [text|json]")

	logging.SetMinimalLevel(logging.CRITICAL)
	vertex.ReadConfigs()

	success := vertex.RunCLITest(*apiName, *serverAddr, *category, *format)
	if !success {
		fmt.Fprintln(os.Stderr, "Tests Failed")
		os.Exit(-1)
	}
	fmt.Println("PASS!")
}
