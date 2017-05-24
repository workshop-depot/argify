package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/comail/colog"
	"github.com/dc0d/argify"
	"github.com/hashicorp/hcl"
	"github.com/urfave/cli"
)

var (
	cnf ConfPile

	// DefaultConfNameHandler generates conf name,
	// conf file name: conf-<app_name>.conf or app.conf
	DefaultConfNameHandler = func() string {
		fp := fmt.Sprintf("conf-%s.conf", DefaultAppNameHandler())
		if _, err := os.Stat(fp); err != nil {
			fp = "app.conf"
		}
		return fp
	}

	// DefaultAppNameHandler generates app name,
	// default is filepath.Base(os.Args[0])
	DefaultAppNameHandler = func() string {
		return filepath.Base(os.Args[0])
	}
)

// LoadHCL loads hcl conf files;
// expected conf file name: DefaultConfNameHandler
func LoadHCL(ptr interface{}, filePath ...string) error {
	var fp string
	if len(filePath) > 0 {
		fp = filePath[0]
	}
	if fp == "" {
		fp = DefaultConfNameHandler()
	}
	cn, err := ioutil.ReadFile(fp)
	if err != nil {
		return err
	}
	err = hcl.Unmarshal(cn, ptr)
	if err != nil {
		return err
	}

	return nil
}

type CouchDBInfo struct {
	URL      string
	User     string
	Password string
	DBName   string
}

// ConfPile sample conf struct
type ConfPile struct {
	Version1 string `name:"v1" usage:"--v1 0.0.1" hidden:"false" envvar:"V1" value:"0.0.1"`
	Version2 string `name:"v2" usage:"--v2 0.0.1" hidden:"false" value:"0.0.1"`

	FBool     bool
	FDuration time.Duration
	FFloat64  float64
	FInt64    int64
	FInt      int `value:"33"`
	FString   string
	FUint     uint   `value:"2"`
	FUint64   uint64 `value:"66"`

	App struct {
		LogDir       string
		PostponeExit time.Duration
		Env          string
		Version1     string
	}

	// one object for each database
	PrimaryDB CouchDBInfo

	Start struct {
		Path     string `name:"path,p" usage:"-p ~/C" hidden:"false" value:"/tmp"`
		Interval int
		Server   struct {
			Port int `value:"8080"`
		}
	}
}

func init() {
	colog.Register()
	colog.SetFormatter(&colog.StdFormatter{Colors: true, Flag: log.Lshortfile})

	if err := LoadHCL(&cnf); err != nil {
		switch err.(type) {
		case *os.PathError:
			log.Println("warn:", err)
		default:
			panic(err)
		}
	}
}

func before(ctx *cli.Context) error {
	return nil
}

func after(ctx *cli.Context) error {
	return nil
}

func appInfo() *cli.App {
	app := cli.NewApp()
	app.Usage = "some app"
	app.Version = "0.0.1"
	app.Author = "dc0d"
	app.Copyright = "dc0d"
	app.Description = fmt.Sprintf("Built Time %v", time.Now())
	app.Name = "gistcli"

	app.Before = before
	app.After = after

	return app
}

func show(v interface{}) {
	js, _ := json.MarshalIndent(v, "", "  ")
	log.Printf("JSON:\n%s", js)
}

func appCommands(app *cli.App) {
	{
		app.Action = func(*cli.Context) error {
			show(&cnf)
			return nil
		}
	}

	{
		c := cli.Command{
			Name:  `start`,
			Usage: `starts app`,
			Action: func(*cli.Context) error {
				log.Printf("%+v", cnf)
				return nil
			},
		}
		c.Subcommands = append(c.Subcommands, cli.Command{
			Name: `server`,
			Action: func(*cli.Context) error {
				log.Printf("%+v", cnf)
				return nil
			},
		})
		app.Commands = append(app.Commands, c)
	}
}

func initApp() *cli.App {
	app := appInfo()
	appCommands(app)

	argify.NewArgify().Build(app, &cnf)

	return app
}

func main() {
	app := initApp()
	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln("error:", err)
	}
}
