package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Handlers map[string][]string `yaml:"handlers"`
	Aliases  map[string]string   `yaml:"aliases"`
}

type Filedata struct {
	Mimetype string `json:"mimetype"`
}

var (
	cfg = &Config{}
)

func readConfig() error {
	fn, ok := os.LookupEnv("FP_CONFIG")
	if !ok {
		u, err := user.Current()
		if err != nil {
			return err
		}

		fn = filepath.Join(u.HomeDir, ".fp.yml")
	}

	f, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer f.Close()

	return yaml.NewDecoder(f).Decode(cfg)
}

func getMimetype(url string) (string, error) {
	c, err := http.Get(url + ".json")
	if err != nil {
		return "", err
	}
	defer c.Body.Close()

	fd := &Filedata{}
	if err := json.NewDecoder(c.Body).Decode(fd); err != nil {
		return "", err
	}

	return fd.Mimetype, nil
}

func getCommand(mt string) ([]string, error) {
	for pattern, command := range cfg.Handlers {
		g, err := glob.Compile(pattern)
		if err != nil {
			return nil, err
		}

		if g.Match(mt) {
			return command, nil
		}
	}

	return nil, fmt.Errorf("no handler found for mimetype: %s", mt)
}

func run(aliasOrUrl string, extra_args []string) error {
	url_ := aliasOrUrl
	for alias, u := range cfg.Aliases {
		if alias == aliasOrUrl {
			url_ = u
			break
		}
	}

	u, err := url.Parse(url_)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(strings.ToLower(u.Scheme), "http") {
		return fmt.Errorf("invalid alias or url: %s", url_)
	}

	mt, err := getMimetype(url_)
	if err != nil {
		return err
	}

	args, err := getCommand(mt)
	if err != nil {
		return err
	}
	args = append(args, extra_args...)
	args = append(args, url_)

	fmt.Printf("Playing: %s\n", url_)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func fatal(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func main() {
	fatal(readConfig())

	if _, ok := os.LookupEnv("COMP_LINE"); ok {
		if len(os.Args) >= 3 {
			for alias := range cfg.Aliases {
				if strings.HasPrefix(alias, os.Args[2]) {
					fmt.Println(alias)
				}
			}
		}
		return
	}

	if len(os.Args) < 2 {
		fatal(errors.New("alias or url required!"))
	}

	fatal(run(os.Args[1], os.Args[2:]))
}
