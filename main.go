package main

import (
	"encoding/json"
	"github.com/dforsyth/jot"
	"gopkg.in/alecthomas/kingpin.v2"
	"log"
	"os"
	"reflect"
	"sort"
	"strings"
)

const DefaultPackage = "generated"

func readInFile(p string) map[string]interface{} {
	obj := make(map[string]interface{})

	f, err := os.Open(p)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	if err := dec.Decode(&obj); err != nil {
		panic(err)
	}

	return obj
}

func main() {
	p := kingpin.Flag("path", "json path").Default(".").String()
	name := kingpin.Flag("name", "model name.").Required().String()
	packageName := kingpin.Flag("package", "output package.").Default(DefaultPackage).String()
	replace := kingpin.Flag("override", "type override.").StringMap()

	kingpin.Parse()

	obj := readInFile(*p)

	lowered := strings.ToLower(*packageName)

	modelSpec := jot.Struct(*name)
	fileSpec := jot.File(lowered).
		AddStruct(modelSpec)

	keys := make([]string, len(obj), len(obj))
	for key := range obj {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		// massage the key
		nkey := strings.Replace(key, "_", " ", -1)
		nkey = strings.Title(nkey)
		nkey = strings.Replace(nkey, " ", "", -1)

		val := obj[key]
		if reflect.TypeOf(val) == nil {
			log.Printf("null type: %s, skipping\n", key)
			continue
		}

		var typ jot.TypeSpec
		if replaceName, ok := (*replace)[key]; ok {
			typ = jot.TypeString(replaceName)
		} else {
			typ = jot.Type(reflect.TypeOf(val))
		}
		modelSpec.AddField(jot.Field(nkey, typ).SetTag("json", key))
		keys = append(keys, key)
	}

	if err := fileSpec.Generate(os.Stdout); err != nil {
		fileSpec.Write(os.Stdout)
		panic(err)
	}
}
