package main

import (
	"encoding/json"
	"fmt"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/user"
	"regexp"
	"strings"
)

func readConfigFile(path string) []byte {
	if buf, err := ioutil.ReadFile(path); err != nil {
		panic(err)
	} else {
		return buf
	}
}

func parseConfig(buffer []byte, json_file bool) map[string]interface{} {
	var config map[string]interface{}
	if json_file {
		if err := json.Unmarshal(buffer, &config); err != nil {
			panic(err)
		}
	} else {
		if err := yaml.Unmarshal(buffer, &config); err != nil {
			panic(err)
		}
	}
	return config
}

func parse(configFileName string) map[string]interface{} {
	var config map[string]interface{}
	if matched, _ := regexp.MatchString("\\.json$", configFileName); matched {
		config = parseConfig(readConfigFile(configFileName), true)
		return config
	} else if matched, err := regexp.MatchString("\\.y.*ml$", configFileName); matched {

		config = parseConfig(readConfigFile(configFileName), false)
		return config
	} else {
		panic(err)
	}

}

func expandPath(path string) string {
	usr, _ := user.Current()
	dir := usr.HomeDir
	if path[:2] == "~/" {
		path = strings.Replace(path, "~", dir, 1)
	}
	return path
}

func find(array []string, value string) bool {
	for _, v := range array {
		if v == value {
			return true
		}
	}
	return false
}

func extractVars(conf map[string]interface{}, section string) []string {
	var ret []string
	if c, ok := conf[section].(map[string]interface{}); ok {
		for group, value := range c {
			if group != "_default" {
				for k, _ := range value.(map[string]interface{}) {
					if !find(ret, k) {
						ret = append(ret, k)
					}
				}
			}
		}
	}
	return ret

}

func generateEraseScript(varNames []string) {
	for _, k := range varNames {
		if k != "_command" {
			fmt.Printf("set -ex %s;\n", k)
		}
	}
}

func generateScript(conf map[string]interface{}, section string, zoneName string) {
	fmt.Printf("set -gx %s_name \"%s\";\n", section, zoneName)
	if z, ok := conf[section]; ok {
		zone := z.(map[string]interface{})
		for k, v := range zone[zoneName].(map[string]interface{}) {
			if k == "_command" {
				fmt.Printf("eval (%s);\n", v)
			} else {
				fmt.Printf("set -gx %s \"%v\";\n", k, v)
			}
		}
	}

}

func main() {
	app := cli.NewApp()
	app.Name = "gosw"
	app.Usage = "generate shell scripts in order to switch env configs"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Value: "~/.settings.json",
			Usage: "config file",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "show",
			Usage: "show config sections or a specific section",
			Action: func(c *cli.Context) error {
				config := parse(expandPath(c.GlobalString("config")))
				if !c.Args().Present() {
					for k, _ := range config {
						fmt.Printf("%s\n", k)
					}
				} else {
					if section, ok := config[c.Args().First()]; ok {
						for k, _ := range section.(map[string]interface{}) {
							if k != "_default" {
								fmt.Printf("%s\n", k)
							}
						}

					} else {
						fmt.Errorf("Unknown config section %s\n", c.Args().First())
					}
				}

				return nil
			},
		},
		{
			Name:  "load",
			Usage: "generate the shell script for loading the config section",
			Action: func(c *cli.Context) error {
				config := parse(expandPath(c.GlobalString("config")))
				if c.Args().Present() {
					if section, ok := config[c.Args().First()]; ok {
						// extract var names from configs
						varNames := extractVars(config, c.Args().First())
						var zoneName = c.Args().Get(1)
						// load default if zoneName is missing
						if zoneName == "" {
							s := section.(map[string]interface{})
							zoneName = s["_default"].(string)
						}

						if _, ok := section.(map[string]interface{})[zoneName]; ok {
							generateEraseScript(varNames)
							generateScript(config, c.Args().First(), zoneName)

						} else {
							fmt.Printf("Can't find %s in %s\n", zoneName, c.Args().First())
						}
					} else {
						fmt.Printf("Can't find %s\n", c.Args().First())
					}
				} else {
					fmt.Printf("Arguments: %+v\n", c.Args())
				}
				return nil
			},
		},
		{
			Name:  "defaults",
			Usage: "Generate the default shell script",
			Action: func(c *cli.Context) error {
				config := parse(expandPath(c.GlobalString("config")))
				for section, c := range config {
					generateEraseScript(extractVars(config, section))
					generateScript(config, section, c.(map[string]interface{})["_default"].(string))
				}
				return nil
			},
		},
	}

	app.Run(os.Args)
}
