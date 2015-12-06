package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/EverythingMe/vertex/swagger"
	_ "github.com/EverythingMe/vertex/vertex-generator/java"
	"github.com/EverythingMe/vertex/vertex-generator/registry"
)

func die(msg string, args ...interface{}) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf(msg, args...))
	os.Exit(-1)
}

func main() {

	genName := flag.String("gen", "java", "Generator name (currently only java supported)")
	swaggerUrl := flag.String("swagger", "", "http URL or file:// URI of a vertex API swagger file. - for stdin")

	flag.Parse()

	if *swaggerUrl == "" {
		die("No swagger url given")
	}

	gen, found := registry.Get(*genName)
	if !found {
		die("Could not find generator %s", *genName)
	}

	var input io.Reader

	switch {
	case strings.HasPrefix(*swaggerUrl, "http"):
		resp, err := http.Get(*swaggerUrl)
		if err != nil {
			die("Error getting swagger from url '%s': %s", *swaggerUrl, err)

		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {

			b, _ := ioutil.ReadAll(resp.Body)
			die("Error reading swagger: status code %d, content: %s", resp.StatusCode, string(b))
		}
		input = resp.Body
	case strings.HasPrefix(*swaggerUrl, "file://"):

		fp, err := os.Open(strings.TrimPrefix(*swaggerUrl, "file://"))
		if err != nil {
			die("Could not open %s: %s", *swaggerUrl, err)
		}
		defer fp.Close()
		input = fp
	case *swaggerUrl == "-":
		input = os.Stdin
	default:
		die("Invalid swagger URI given: %s", *swaggerUrl)
	}

	var swapi swagger.API

	if err := json.NewDecoder(input).Decode(&swapi); err != nil {
		die("Error decoding swagger: %s", err)

	}

	b, err := gen.Generate(&swapi)
	if err != nil {
		die("Error decoding swagger: %s", err)
	}

	fmt.Println(string(b))

}
