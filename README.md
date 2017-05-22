# argify
**convention based, argument/flag for awesome [urfave/cli](https://github.com/urfave/cli)**

# status

Not thoroughly tested, for anyone interested!
> in less fun words; WIP, tests, etc etc

# quick peek

We can define a config struct as (_sample_):

```go
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

    // embedded structs would produce arguments with names like --<field name>-<embedded field name>
	App struct {
		LogDir       string
		PostponeExit time.Duration
		Env          string
		Version1     string
	}

	// one object for each database
	PrimaryDB CouchDBInfo

    // we have a command named 'start', so these would get added as it's arguments
	Start struct {
		Path     string `name:"path,p" usage:"-p ~/C" hidden:"false" value:"/tmp"`
		Interval int
	}
}
```

And then load it from a file. It's fields would get bind to cli's flags. If there is a non-zero vlue, it will be used as default value. Also values can get defined using Go struct tags.

Having an instance of `*cli.App`, after commands and subcommands are defined all we have to do is `argify.NewArgify().Build(app, &cnf)`.