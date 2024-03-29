package cobrax

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type FlagTemplateFun func(flagType, flagName, value string) string
type FlagsGet func(*cobra.Command) *pflag.FlagSet

var (
	GetFlags FlagsGet = func(command *cobra.Command) *pflag.FlagSet {
		return command.Flags()
	}
	PersistentFlags FlagsGet = func(command *cobra.Command) *pflag.FlagSet {
		return command.PersistentFlags()
	}
	LocalFlags FlagsGet = func(command *cobra.Command) *pflag.FlagSet {
		return command.LocalFlags()
	}
)

func Flags(cmd *cobra.Command, main interface{}, prefix, envPrefix string, tempFn ...FlagTemplateFun) error {
	return FlagsWith(cmd, GetFlags, main, prefix, envPrefix, tempFn...)
}

func FlagsWith(cmd *cobra.Command, fn FlagsGet, main interface{}, prefix, envPrefix string, tempFn ...FlagTemplateFun) error {
	if main == nil || reflect.ValueOf(main).IsZero() || reflect.ValueOf(main).IsNil() {
		return fmt.Errorf("main is zero Value")
	}
	typ := reflect.TypeOf(main)
	if typ.Kind() != reflect.Ptr {
		return fmt.Errorf("value must be pointer to struct, but is %s", typ.Kind().String())
	}

	mainVal := reflect.ValueOf(main).Elem()
	mainTyp := mainVal.Type()
	if mainTyp.Kind() != reflect.Struct {
		return fmt.Errorf("value must be pointer to struct, but is pointer to %s", typ.Kind().String())
	}
	return setFlags(cmd, fn(cmd), main, prefix, envPrefix, tempFn...)
}

func tag(field reflect.StructField, names ...string) string {
	for _, name := range names {
		if value, ok := field.Tag.Lookup(name); ok {
			return value
		}
	}
	return ""
}

func templateValue(flagType, flagName, value string, tempFn ...FlagTemplateFun) string {
	if value == "" || !strings.Contains(value, "{{") {
		return value
	}
	if len(tempFn) > 0 {
		for _, fun := range tempFn {
			if newV := fun(flagType, flagName, value); newV != value && newV != "" {
				return newV
			}
		}
	}
	return value
}

func getValue(flagSet *pflag.FlagSet, name, env string) func() error {
	return func() error {
		flag := flagSet.Lookup(name)
		if flag.Changed {
			return nil
		}
		value, has := os.LookupEnv(env)
		if has {
			return flagSet.Set(name, value)
		}
		return nil
	}
}

func setFlags(cmd *cobra.Command, flags *pflag.FlagSet, main interface{}, prefix, envPrefix string, tempFn ...FlagTemplateFun) error {
	//flags /* flag.FlagSet*/ := cmd.PersistentFlags()

	mainVal := reflect.ValueOf(main).Elem()
	mainTyp := mainVal.Type()
	fns := make([]func() error, 0)

	for i := 0; i < mainTyp.NumField(); i++ {
		fieldType := mainTyp.Field(i)
		fieldVal := mainVal.Field(i)

		//An identifier may be exported to permit access to it from another package. An identifier is exported if both:
		//the first character of the identifier's name is a Unicode upper case letter (Unicode class "Lu"); and
		//the identifier is declared in the package block or it is a field name or method name.
		//All other identifiers are not exported.
		if fieldType.PkgPath != "" {
			continue
		}

		flagName, has := fieldType.Tag.Lookup("flag")
		if flagName == "-" {
			continue
		} else if !has && flagName == "" {
			flagName = camel2Case(fieldType.Name, '-')
		}

		if prefix != "" {
			if flagName == "" {
				flagName = prefix
			} else {
				flagName = prefix + "." + flagName
			}
		}

		shorthand := tag(fieldType, "short")
		env := tag(fieldType, "env")
		if env == "-" {
			env = ""
		} else if env == "" {
			env = envName(flagName)
		}
		if envPrefix != "" {
			env = envPrefix + "_" + env
		}
		help := templateValue("help", flagName, tag(fieldType, "help", "h"), tempFn...)
		if env != "" {
			help += " (env: " + env + ") "
		}
		defValue := templateValue("def", flagName, tag(fieldType, "def"), tempFn...)

		if fieldVal.Type().Kind() == reflect.Ptr {
			if fieldVal.IsNil() {
				fieldVal.Set(reflect.New(fieldType.Type.Elem()))
			}
			fieldVal = fieldVal.Elem()
		}

		switch fieldVal.Interface().(type) {
		case time.Duration:
			p := fieldVal.Addr().Interface().(*time.Duration)
			value := time.Duration(fieldVal.Int())
			if defValue != "" {
				if d, err := time.ParseDuration(defValue); err == nil {
					value = d
				}
			}
			flags.DurationVarP(p, flagName, shorthand, value, help)
			goto ENV
		case net.IP:
			p := fieldVal.Addr().Interface().(*net.IP)
			value := net.IP(fieldVal.Bytes())
			if defValue != "" {
				value = net.ParseIP(defValue)
			}
			flags.IPVarP(p, flagName, shorthand, value, help)
			goto ENV
		case []net.IP:
			p := fieldVal.Addr().Interface().(*[]net.IP)
			flags.IPSliceVarP(p, flagName, shorthand, *p, help)
			goto ENV
		case []string:
			p := fieldVal.Addr().Interface().(*[]string)
			value := *p
			if defValue != "" {
				value = strings.Split(defValue, ",")
			}
			if value == nil {
				value = make([]string, 0)
			}
			flags.StringSliceVarP(p, flagName, shorthand, value, help)
			goto ENV
		}

		// now check basic kinds
		switch fieldVal.Kind() {
		case reflect.String:
			p := fieldVal.Addr().Interface().(*string)
			value := *p
			if defValue != "" {
				value = defValue
			}
			flags.StringVarP(p, flagName, shorthand, value, help)
		case reflect.Bool:
			p := fieldVal.Addr().Interface().(*bool)
			value := fieldVal.Bool()
			if defValue != "" {
				value, _ = strconv.ParseBool(defValue)
			}
			flags.BoolVarP(p, flagName, shorthand, value, help)
		case reflect.Int:
			p := fieldVal.Addr().Interface().(*int)
			val := int(fieldVal.Int())
			if defValue != "" {
				val, _ = strconv.Atoi(defValue)
			}
			flags.IntVarP(p, flagName, shorthand, val, help)
		case reflect.Int64:
			p := fieldVal.Addr().Interface().(*int64)
			val := *p
			if defValue != "" {
				val, _ = strconv.ParseInt(defValue, 10, 64)
			}
			flags.Int64VarP(p, flagName, shorthand, val, help)
		case reflect.Float64:
			p := fieldVal.Addr().Interface().(*float64)
			value := *p
			if defValue != "" {
				value, _ = strconv.ParseFloat(defValue, 64)
			}
			flags.Float64VarP(p, flagName, shorthand, value, help)
		case reflect.Uint:
			p := fieldVal.Addr().Interface().(*uint)
			val := *p
			if defValue != "" {
				v, _ := strconv.ParseUint(defValue, 10, 64)
				val = uint(v)
			}
			flags.UintVarP(p, flagName, shorthand, val, help)
		case reflect.Uint64:
			p := fieldVal.Addr().Interface().(*uint64)
			val := *p
			if defValue != "" {
				val, _ = strconv.ParseUint(defValue, 10, 64)
			}
			flags.Uint64VarP(p, flagName, shorthand, val, help)
		case reflect.Slice:
			switch fieldType.Type.Elem().Kind() {
			case reflect.Int:
				p := fieldVal.Addr().Interface().(*[]int)
				val := *p
				if defValue != "" {
					sVal := strings.Split(defValue, ",")
					for i, s := range sVal {
						if v, err := strconv.Atoi(strings.TrimSpace(s)); err != nil {
							return err
						} else {
							val[i] = v
						}
					}
				}
				if val == nil {
					val = make([]int, 0)
				}
				flags.IntSliceVarP(p, flagName, shorthand, val, help)
			case reflect.String:
				p := fieldVal.Addr().Interface().(*[]string)
				val := *p
				if defValue != "" {
					val = strings.Split(defValue, ",")
					for i, s := range val {
						val[i] = strings.TrimSpace(s)
					}
				}
				if val == nil {
					val = make([]string, 0)
				}
				flags.StringSliceVarP(p, flagName, shorthand, val, help)
			default:
				return fmt.Errorf("encountered unsupported slice type/kind: %s", fieldType.Type.Name())
			}
		case reflect.Struct:
			newPrefix := flagName
			if strings.HasSuffix(flagName, "!embed") {
				newPrefix = prefix
			}
			err := setFlags(cmd, flags, fieldVal.Addr().Interface(), newPrefix, envPrefix, tempFn...)
			if err != nil {
				return err
			}
			continue
		case reflect.Map:
			p, match := fieldVal.Addr().Interface().(*map[string]string)
			if !match {
				return fmt.Errorf("encountered unsupported fieldType type/kind: %v", fieldType.Type.String())
			}
			val := *p
			if defValue != "" {
				val = make(map[string]string)
				for _, pair := range strings.Split(defValue, ",") {
					kv := strings.SplitN(strings.TrimSpace(pair), "=", 2)
					if len(kv) == 1 {
						val[strings.TrimSpace(kv[0])] = ""
					} else {
						val[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
					}
				}
			}
			flags.StringToStringVarP(p, flagName, shorthand, val, help)
		default:
			return fmt.Errorf("encountered unsupported  type/kind: %s", fieldType.Type.String())
		}
	ENV:
		if env != "" {
			fns = append(fns, getValue(flags, flagName, env))
		}
	}

	ppre := cmd.PersistentPreRunE
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		for _, envFn := range fns {
			if err := envFn(); err != nil {
				return err
			}
		}
		if ppre != nil {
			return ppre(cmd, args)
		}
		return nil
	}
	return nil
}
