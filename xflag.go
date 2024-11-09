package xflag

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type flag struct {
	name  string
	value string
}

type userInput struct {
	binName   *string
	command   *string
	arguments []string
	flags     []flag
}

type argumentDef struct {
	field    string
	name     string
	usage    string
	datatype string
}

type flagDef struct {
	field        string
	usage        string
	defaultValue *string
	datatype     string
}

type commandDef struct {
	value     reflect.Value
	name      *string
	usage     *string
	arguments []argumentDef
	flags     map[string]flagDef
}

func Parse(commands []interface{}, args []string) (interface{}, error) {
	commandDefs := parseCommandDefs(commands)

	userInput, err := parseUserInput(args)
	if err != nil {
		return nil, err
	}

	if userInput.command == nil {
		return nil, errors.New(helpOverview(commandDefs))
	}

	var def *commandDef
	for _, cmd := range commandDefs {
		if *cmd.name == *userInput.command {
			def = &cmd
			break
		}
	}

	if def == nil {
		return nil, fmt.Errorf("%s\nunrecognized command \"%s\"\n", helpOverview(commandDefs), *userInput.command)
	}

	if len(userInput.arguments) > len(def.arguments) {
		return nil, fmt.Errorf("%s\ntoo many arguments given\n", helpCommandDetails(def))
	}

	for i, arg := range def.arguments {
		var errtext strings.Builder
		if i < len(userInput.arguments) {
			errtext.WriteString(fmt.Sprintf("  [ok] found argument <%s>\n", arg.name))
		}
		if i >= len(userInput.arguments) {
			if i > 0 {
				errtext.WriteString(fmt.Sprintf("[fail] missing argument <%s>\n", arg.name))
				return nil, fmt.Errorf("%s\n%s", helpCommandDetails(def), errtext.String())
			} else {
				return nil, fmt.Errorf("%s", helpCommandDetails(def))
			}
		}
	}

	t := def.value.Type()
	instance := reflect.New(t)

	// set arguments
	for i, argument := range def.arguments {
		err = setField(instance, argument.field, userInput.arguments[i])
		if err != nil {
			return nil, fmt.Errorf("unable to set argument field \"%s\" to \"%s\"", argument.field, userInput.arguments[i])
		}
	}

	// apply defaults
	for _, fdef := range def.flags {
		if fdef.defaultValue == nil {
			continue
		}

		err = setField(instance, fdef.field, *fdef.defaultValue)
		if err != nil {
			return nil, fmt.Errorf("unable to set default flag field \"%s\" to \"%s\"", fdef.field, *fdef.defaultValue)
		}
	}

	// set flags
	for _, flag := range userInput.flags {
		fdef, ok := def.flags[flag.name]
		if !ok {
			return nil, fmt.Errorf("unrecognized flag \"%s\"", flag.name)
		}

		err = setField(instance, fdef.field, flag.value)
		if err != nil {
			return nil, fmt.Errorf("unable to set flag field \"%s\" to \"%s\"", fdef.field, flag.value)
		}
	}

	result := instance.Interface()
	return result, nil
}

func parseCommandDefs(commands []interface{}) []commandDef {
	var cmds []commandDef

	for _, command := range commands {
		v := reflect.ValueOf(command)

		if v.Kind() == reflect.Ptr {
			panic("do not pass in pointers\n")
		}

		var cmd commandDef
		cmd.value = v
		cmd.flags = make(map[string]flagDef)

		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			tag := field.Tag.Get("xflag")
			parts := strings.Split(tag, "|")

			if field.Name == "XFlag" {
				if len(parts) != 2 {
					panic("command missing usage\n")
				}

				cmd.name = &parts[0]
				cmd.usage = &parts[1]
				continue
			}

			if field.Type.Kind() == reflect.Ptr {
				var f flagDef
				name := toKebabCase(field.Name)
				if len(parts) == 2 {
					f.field = field.Name
					f.defaultValue = &parts[0]
					f.usage = parts[1]
				} else if len(parts) == 1 {
					f.field = field.Name
					f.usage = parts[0]
				} else {
					panic("flag missing usage\n")
				}
				f.datatype = field.Type.String()
				cmd.flags[name] = f
			} else {
				var a argumentDef
				if len(parts) == 1 {
					a.field = field.Name
					a.name = strings.ToLower(field.Name)
					a.usage = parts[0]
				} else {
					panic("argument lacks exactly one field\n")
				}
				a.datatype = field.Type.String()
				cmd.arguments = append(cmd.arguments, a)
			}
		}

		if cmd.name == nil || cmd.usage == nil {
			panic("command missing name/usage\n")
		}

		cmds = append(cmds, cmd)
	}

	return cmds
}

func parseUserInput(args []string) (*userInput, error) {
	var input userInput

	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "-") || args[i] == "help" {
			name := strings.TrimPrefix(args[i], "-")
			if name == "help" || name == "-help" || args[i] == "-h" {
				continue
			}
			if i+1 >= len(args) {
				return nil, fmt.Errorf("missing value for flag \"%s\"", name)
			}
			f := flag{name, args[i+1]}
			i++
			input.flags = append(input.flags, f)
		} else {
			input.arguments = append(input.arguments, args[i])
		}
	}

	if len(input.arguments) >= 1 {
		input.binName = &input.arguments[0]

		if len(input.arguments) > 1 {
			input.command = &input.arguments[1]
			input.arguments = input.arguments[2:]
		}
	}

	return &input, nil
}

func toKebabCase(input string) string {
	var result []rune
	for i, r := range input {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '-')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}

func setField(instance reflect.Value, name, value string) error {
	field := instance.Elem().FieldByName(name)
	if !field.IsValid() || !field.CanSet() {
		return fmt.Errorf("cannot set field %s", name)
	}

	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		field = field.Elem()
	}

	switch field.Kind() {
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolValue)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) { // Special case for time.Duration
			durationValue, err := time.ParseDuration(value)
			if err != nil {
				return err
			}
			field.SetInt(int64(durationValue))
		} else {
			intValue, err := strconv.ParseInt(value, 10, field.Type().Bits())
			if err != nil {
				return err
			}
			field.SetInt(intValue)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		uintValue, err := strconv.ParseUint(value, 10, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetUint(uintValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetFloat(floatValue)
	case reflect.String:
		field.SetString(value)
	default:
		return fmt.Errorf("unsupported kind %s", field.Kind())
	}
	return nil
}

func GetUsage(commands []interface{}) string {
	commandDefs := parseCommandDefs(commands)
	return helpOverview(commandDefs)
}

func helpOverview(defs []commandDef) string {
	var rows []row

	for _, def := range defs {
		rows = append(rows, row{*def.name, *def.usage, nil, nil})
	}

	return "\n" + formatTable(rows)
}

func helpCommandDetails(def *commandDef) string {
	var rows []row
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("\nusage: %s ", *def.name))
	for _, a := range def.arguments {
		name := "<" + a.name + ">"
		usage := a.usage
		datatype := a.datatype
		rows = append(rows, row{name, usage, &datatype, nil})
		buf.WriteString(fmt.Sprintf("%s ", name))
	}
	for name, f := range def.flags {
		name = "[-" + name + "]"
		usage := f.usage
		datatype := f.datatype
		rows = append(rows, row{name, usage, &datatype, f.defaultValue})
		buf.WriteString(fmt.Sprintf("%s ", name))
	}

	buf.WriteString("\n\n")
	return buf.String() + formatTable(rows)
}

type row struct {
	key          string
	value        string
	datatype     *string
	defaultValue *string
}

func formatTable(rows []row) string {
	var buf strings.Builder
	var longestKey int
	var longestValue int
	var longestDatatype int

	for _, row := range rows {
		if l := len(row.key); longestKey < l {
			longestKey = l
		}
		if l := len(row.value); longestValue < l {
			longestValue = l
		}
		if row.datatype != nil && longestDatatype < len(*row.datatype) {
			longestDatatype = len(*row.datatype)
		}
	}

	for _, row := range rows {
		space := 4 + longestKey - len(row.key)
		tabs := strings.Repeat(" ", space)
		fmtstr := fmt.Sprintf("%%s%%s.......%%-%ds", longestValue+4)
		buf.WriteString(fmt.Sprintf(fmtstr, tabs, row.key, row.value))
		if row.datatype != nil {
			dt := strings.Trim(*row.datatype, "*")
			dt = strings.TrimPrefix(dt, "time.")
			dt = strings.ToLower(dt)
			pad := longestDatatype - len(dt)
			buf.WriteString(fmt.Sprintf("%s", dt))
			buf.WriteString(strings.Repeat(" ", pad))
			if row.defaultValue != nil {
				buf.WriteString(fmt.Sprintf("  (default: \"%s\")", *row.defaultValue))
			}
		}
		buf.WriteString("\n")
	}

	buf.WriteString("\n")
	return buf.String()
}
