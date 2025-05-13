package config_test

import (
	"testing"

	"github.com/xakepp35/anygate/config"
	"gopkg.in/yaml.v3"
)

func TestRoot(t *testing.T) {

	const yml = `
routes:
  /test/1: 2
  GET /test2: http://api/v2
  GET,POST /test2: http://api/v2
  GET POST /test2: http://api/v2
`

	var cfg config.Root

	err := yaml.Unmarshal([]byte(yml), &cfg)
	if err != nil {
		t.Fatalf("err %v", err)
	}
	t.Fatalf("%v", cfg)

}
