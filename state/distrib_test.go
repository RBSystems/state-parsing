package state

import (
	"log"
	"testing"
)

func TestMapCopy(t *testing.T) {
	//a
	a := make(map[string]interface{})
	a["a"] = "a"
	a["b"] = make(map[string]interface{})
	a["c"] = "c"
	a["d"] = make(map[string]interface{})

	a["b"].(map[string]interface{})["1"] = "b1"
	a["b"].(map[string]interface{})["2"] = make(map[string]interface{})
	a["b"].(map[string]interface{})["3"] = "b3"
	a["b"].(map[string]interface{})["4"] = make(map[string]interface{})

	a["b"].(map[string]interface{})["2"].(map[string]interface{})["I"] = "b2I"
	a["b"].(map[string]interface{})["2"].(map[string]interface{})["II"] = "b2II"

	a["d"].(map[string]interface{})["1"] = "d1"
	a["d"].(map[string]interface{})["2"] = make(map[string]interface{})
	a["d"].(map[string]interface{})["3"] = "d3"

	a["d"].(map[string]interface{})["2"].(map[string]interface{})["I"] = "d2I"
	a["d"].(map[string]interface{})["2"].(map[string]interface{})["II"] = "d2II"

	b := make(map[string]interface{})

	b["a"] = "x"
	b["b"] = make(map[string]interface{})
	b["c"] = "c"
	b["d"] = make(map[string]interface{})
	b["e"] = "y"

	b["b"].(map[string]interface{})["1"] = "b1"
	b["b"].(map[string]interface{})["2"] = make(map[string]interface{})
	b["b"].(map[string]interface{})["3"] = "b3"
	b["b"].(map[string]interface{})["4"] = make(map[string]interface{})

	b["b"].(map[string]interface{})["2"].(map[string]interface{})["II"] = "c2II"

	b["d"].(map[string]interface{})["1"] = "z"
	b["d"].(map[string]interface{})["4"] = make(map[string]interface{})
	b["d"].(map[string]interface{})["5"] = "aa"

	log.Printf("Pre B: %v", b)

	replaceMapValues(&a, &b)

	log.Printf("A: %v\n\n", a)
	log.Printf("B: %v", b)
}
