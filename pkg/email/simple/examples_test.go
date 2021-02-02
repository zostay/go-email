package simple_test

import (
	"io/ioutil"
	"strings"

	"github.com/zostay/go-email/pkg/email/simple"
)

func Example() {
	msg, err := ioutil.ReadFile("input.msg")
	if err != nil {
		panic(err)
	}

	m, err := simple.Parse(msg)
	if err != nil {
		panic(err)
	}

	if kw := m.HeaderGet("Keywords"); kw != "" {
		kws := strings.Split(kw, ",")
		for _, k := range kws {
			if strings.TrimSpace(k) == "Snuffle" {
				kw += ", Upagus"
				_ = m.HeaderSet("Keywords", kw)
				_ = ioutil.WriteFile("output.msg", m.Bytes(), 0644)
			}
		}
	}
}
