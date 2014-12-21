package fuzz

import (
	"log"
	"regexp"
	"regexp/syntax"
	"testing"
)

const testCount = 10

func TestGen(t *testing.T) {
	in := []string{
		`aaa`,
		`a+`,
		`abc*`,
		`a|c`,
		`(foo|bar){1,10}`,
		`[a-zA-Z]{1,3}`,
		`foo(bar)?buz`,
		`[[:alpha:]]{4}`,
	}
	for _, pattern := range in {
		g, err := NewGenerator(pattern, syntax.Perl)
		if err != nil {
			t.Errorf("%v", err)
		}
		result := make([]string, testCount)
		for i := 0; i < testCount; i++ {
			for {
				r, err := g.Gen()
				if err != nil {
					log.Printf("%v -> %v: min %v max %v\n", r, err, g.min, g.max)
				}
				result[i] = r
				break
			}
		}

		for _, r := range result {
			log.Printf("%v -> %v\n", pattern, r)
			ok, err := regexp.MatchString(pattern, r)
			if err != nil {
				t.Errorf("%v", err)
			}
			if !ok {
				t.Errorf("pattern: %v, generated: %v", pattern, r)
			}
		}
	}
}
