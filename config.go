package vertex

import (
	"github.com/EverythingMe/gofigure"
	"github.com/EverythingMe/gofigure/autoflag"
	"github.com/dvirsky/go-pylog/logging"

	"gopkg.in/yaml.v2"
)

type serverConfig struct {
	// Listening address for the server, e.g. ":8080"
	ListenAddr string `yaml:"listen"`

	// Should we allow non http access to the API? use only on dev machines
	AllowInsecure bool `yaml:"allow_insecure"`

	// The location of the console UI html files on the local machine
	ConsoleFilesPath string `yaml:"console_files_path"`

	// Minimal logging level [DEBUG | INFO | WARN | ERROR | CRITICAL]
	LoggingLevel string `yaml:"logging_level"`

	// Disconnect idle clients after T seconds
	ClientTimeout int `yaml:"client_timeout_sec"`
}

// General-purpose to just protect some urls
type authConfig struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type confType struct {
	Server     serverConfig           `yaml:"server"`
	Auth       authConfig             `yaml:"auth"`
	APIConfigs map[string]interface{} `yaml:"apis"`
}

var Config = struct {
	Server     serverConfig           `yaml:"server"`
	Auth       authConfig             `yaml:"auth"`
	APIConfigs map[string]interface{} `yaml:"apis,flow"`

	apiconfs map[string]interface{}
}{
	Server: serverConfig{
		ListenAddr:       ":9944",
		AllowInsecure:    false,
		ConsoleFilesPath: "../console",
		LoggingLevel:     "INFO",
		ClientTimeout:    60,
	},

	Auth: authConfig{
		User:     "vertext",
		Password: "xetrev",
	},

	APIConfigs: make(map[string]interface{}),

	apiconfs: make(map[string]interface{}),
}

// registerAPIConfig registers the configurations for a specific api, under the path of /apis/<api_name>. e.g
//	apis:
//		myApi:
//			foo: bar
func registerAPIConfig(name string, conf interface{}) {
	Config.apiconfs[name] = conf
}

func ReadConfigs() error {

	logging.Info("ReadING configs: %#v", &Config)
	if err := autoflag.Load(gofigure.DefaultLoader, &Config); err != nil {
		logging.Error("Error loading configs: %v", err)
		return err
	}
	logging.Info("Read configs: %#v", &Config)

	for k, m := range Config.APIConfigs {

		if conf, found := Config.apiconfs[k]; found && conf != nil {

			b, err := yaml.Marshal(m)
			if err == nil {

				if err := yaml.Unmarshal(b, conf); err != nil {
					logging.Error("Error reading config for API %s: %s", k, err)
				} else {
					logging.Debug("Unmarshaled API config for %s: %#v", k, conf)
				}

			} else {

				logging.Error("Error marshalling config for API %s: %s", k, err)

			}
		} else {
			logging.Warning("API Section %s in config file not registered with server", k)
		}

	}

	return nil

}
