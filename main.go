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
						var varNames []string
						// extract var names from configs
						for key, values := range section.(map[string]interface{}) {
							if key != "_default" {
								for k, _ := range values.(map[string]interface{}) {
									if !find(varNames, k) {
										varNames = append(varNames, k)
									}
								}
							}

						}
						zoneName := c.Args().Get(1)
						if zone, ok := section.(map[string]interface{})[zoneName]; ok {
							for _, i := range varNames {
								fmt.Printf("set -ex %s;\n", i)
							}
							fmt.Printf("set -gx %s_name \"%s\";\n", c.Args().First(), zoneName)
							for k, v := range zone.(map[string]interface{}) {
								if k == "_command" {
									fmt.Printf("eval (%s);\n", v)
								} else {
									fmt.Printf("set -gx %s \"%s\";\n", k, v)
								}
							}

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
	}

	app.Run(os.Args)
}
