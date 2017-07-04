package argify

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dc0d/inflect"
	"github.com/dc0d/inflect/inflectlab"
	"github.com/urfave/cli"
)

// Argify adds arguments based on fields of a struct and binds them too
type Argify struct {
	appName string
}

// NewArgify creates an instance of Argify
func NewArgify() *Argify {
	return &Argify{}
}

func (b *Argify) extractSharedParts(
	prefix, commandName string,
	k1 string,
	v1 inflectlab.Field) (name, usage, envVar string, hidden bool) {
	name = strings.ToLower(k1)
	if prefix != "" {
		name = strings.ToLower(prefix) + "-" + name
	}
	if v, ok := v1.Tags["name"]; ok {
		name = v
	}
	usage = ""
	if v, ok := v1.Tags["usage"]; ok {
		usage = v
	}
	envVar = ""
	if v, ok := v1.Tags["envvar"]; ok {
		envVar = v
	}
	if envVar == "" {
		envName := name
		envParts := strings.Split(envName, ",")
		if len(envParts) > 1 {
			envName = envParts[0]
			if len(envParts[1]) > len(envName) {
				envName = envParts[1]
			}
		}
		if commandName == "" {
			envVar = strings.Join([]string{b.appName, envName}, "_")
		} else {
			envVar = strings.Join([]string{b.appName, commandName, envName}, "_")
		}
		envVar = strings.Trim(envVar, "_")
		envVar = strings.Replace(envVar, "-", "_", -1)
		envVar = strings.ToUpper(envVar)
	}
	hidden = false
	if v, ok := v1.Tags["hidden"]; ok {
		if b, err := strconv.ParseBool(v); err == nil {
			hidden = b
		}
	}

	return
}

func (b *Argify) setSharedParts(f cli.Flag, name, usage, envVar string, hidden bool) {
	if err := inflect.Set(f, "Name", name); err != nil {
		// TODO:
	}
	if err := inflect.Set(f, "Usage", usage); err != nil {
		// TODO:
	}
	if envVar != "-" {
		if err := inflect.Set(f, "EnvVar", envVar); err != nil {
			// TODO:
		}
	}
	if err := inflect.Set(f, "Hidden", hidden); err != nil {
		// TODO:
	}
}

func (b *Argify) process(
	prefix, commandName string,
	fields map[string]inflectlab.Field,
	commands *[]cli.Command,
	flags *[]cli.Flag) {
	var keys []string
	for k1 := range fields {
		keys = append(keys, k1)
	}
	sort.Strings(keys)

NEXT_FIELD:
	for _, k1 := range keys {
		v1 := fields[k1]
		for kcmd := range *commands {
			if strings.EqualFold((*commands)[kcmd].Name, k1) {
				var subCommands []cli.Command = (*commands)[kcmd].Subcommands
				var cmdFlags = (*commands)[kcmd].Flags
				if cmdFlags == nil {
					cmdFlags = []cli.Flag{}
				}
				b.process("", (*commands)[kcmd].Name, v1.Children, &subCommands, &cmdFlags)
				(*commands)[kcmd].Subcommands = subCommands
				(*commands)[kcmd].Flags = cmdFlags

				continue NEXT_FIELD
			}
		}

		var f cli.Flag

		//

		defaultValue := reflect.Zero(v1.Ptr.Type()).Interface()
		var tagString string
		if v, ok := v1.Tags["value"]; ok && v != fmt.Sprintf("%v", defaultValue) {
			tagString = v
		}
		var tagValue interface{}

		switch v1.Ptr.Type() {
		case reflect.TypeOf(true):
			f = &cli.BoolFlag{}
			if tagString != "" {
				if vp, err := strconv.ParseBool(tagString); err == nil {
					tagValue = vp
				}
			}
		case reflect.TypeOf(""):
			f = &cli.StringFlag{}
			if tagString != "" {
				tagValue = tagString
			}
		case reflect.TypeOf(time.Second):
			f = &cli.DurationFlag{}
			if tagString != "" {
				if vp, err := time.ParseDuration(tagString); err == nil {
					tagValue = vp
				}
			}
		case reflect.TypeOf(float64(1.0)):
			f = &cli.Float64Flag{}
			if tagString != "" {
				if vp, err := strconv.ParseFloat(tagString, 64); err == nil {
					tagValue = vp
				}
			}
		case reflect.TypeOf(int64(1)):
			f = &cli.Int64Flag{}
			if tagString != "" {
				if vp, err := strconv.ParseInt(tagString, 10, 64); err == nil {
					tagValue = vp
				}
			}
		case reflect.TypeOf(int(1)):
			f = &cli.IntFlag{}
			if tagString != "" {
				if vp, err := strconv.ParseInt(tagString, 10, 64); err == nil {
					tagValue = int(vp)
				}
			}
		case reflect.TypeOf(uint(1)):
			f = &cli.UintFlag{}
			if tagString != "" {
				if vp, err := strconv.ParseInt(tagString, 10, 64); err == nil {
					tagValue = uint(vp)
				}
			}
		case reflect.TypeOf(uint64(1)):
			f = &cli.Uint64Flag{}
			if tagString != "" {
				if vp, err := strconv.ParseInt(tagString, 10, 64); err == nil {
					tagValue = uint64(vp)
				}
			}
		default:
			var cmdFlags = *flags
			if cmdFlags == nil {
				cmdFlags = []cli.Flag{}
			}
			var subCommands []cli.Command
			b.process(k1, "", v1.Children, &subCommands, &cmdFlags)
			*flags = cmdFlags
			continue
		}

		//

		if v1.Ptr.Interface() != defaultValue {
			if err := inflect.Set(f, "Value", v1.Ptr.Interface()); err != nil &&
				err.Error() != "field does not exist" {
				// TODO:
			}
		} else if tagString != "" {
			if err := inflect.Set(f, "Value", tagValue); err != nil &&
				err.Error() != "field does not exist" {
				// TODO:
			}
		}

		name, usage, envVar, hidden := b.extractSharedParts(prefix, commandName, k1, v1)
		b.setSharedParts(f, name, usage, envVar, hidden)
		if err := inflect.Set(f, "Destination", v1.Ptr.Addr().Interface()); err != nil {
			// TODO:
		}

		*flags = append(*flags, f)
	}
}

// Build builds the arguments/flags
func (b *Argify) Build(app *cli.App, confPtr interface{}) error {
	fields, err := inflectlab.GetFields(confPtr)
	if err != nil {
		return err
	}

	b.appName = app.Name
	b.process("", "", fields, &app.Commands, &app.Flags)

	return nil
}
