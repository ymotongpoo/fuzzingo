package fuzz

import (
	"errors"
	"log"
	"math/rand"
	"regexp/syntax"
	"time"
)

var (
	ErrTooFewRepeat  = errors.New("Counted too few repeat.")
	ErrTooManyRepeat = errors.New("Counted too many repeat.")
)

// Generator holds base regular expression pattern .
type Generator struct {
	pattern  string
	flags    syntax.Flags
	prog     *syntax.Prog
	min, max int
}

// NewGenerator returns Generator.
func NewGenerator(pattern string, flags syntax.Flags) (*Generator, error) {
	re, err := syntax.Parse(pattern, flags)
	if err != nil {
		return nil, err
	}
	min := re.Min
	max := re.Max
	re = re.Simplify()
	prog, err := syntax.Compile(re)
	if err != nil {
		return nil, err
	}
	gen := &Generator{
		pattern: pattern,
		flags:   flags,
		prog:    prog,
		min:     min,
		max:     max,
	}
	return gen, nil
}

// Gen creates random string matching pattern.
// TODO(ymotongpoo): replace internal implementation gen with better one.
func (g *Generator) Gen() (string, error) {
	return g.gen()
}

// gen randomly shifts in NFA created from pattern.
// TODO(ymotongpoo): this is test implementation.
func (g *Generator) gen() (string, error) {
	inst := g.prog.Inst
	pc := uint32(g.prog.Start)
	i := inst[pc]
	result := []rune{}
	cap := []uint32{}

	for {
		switch i.Op {
		default:
			log.Fatalf("%v: %v", i.Op, "bad operation")
		case syntax.InstFail:
			// nothing
		case syntax.InstRune:
			r := randRune(i.Rune)
			result = append(result, r)
			pc = i.Out
			i = inst[pc]
		case syntax.InstRune1:
			result = append(result, i.Rune[0])
			pc = i.Out
			i = inst[pc]
		case syntax.InstAlt:
			pc = randPath(i.Out, i.Arg, cap)
			i = inst[pc]
		case syntax.InstCapture:
			cap = append(cap, pc)
			if len(cap) > (g.max+1)*2 {
				return string(result), ErrTooManyRepeat
			}
			pc = randPath(i.Out, i.Arg, cap)
			i = inst[pc]
		case syntax.InstMatch:
			if g.prog.NumCap > 2 && len(cap) < g.min*2 {
				return string(result), ErrTooFewRepeat
			}
			return string(result), nil
		}
	}
}

// randPath must be called when there are alternative paths.
func randPath(out, arg uint32, cap []uint32) uint32 {
	rand.Seed(time.Now().UnixNano())
	if rand.Intn(356)%2 == 0 {
		if len(cap) > 0 && out > cap[len(cap)-1] {
			return out
		}
		return arg
	}
	if len(cap) > 0 && arg > cap[len(cap)-1] {
		return arg
	}
	return out
}

// randAscii returns random ascii character between min and max.
// TODO(ymotongpoo): confirm if generated rune is valid in the case that it is in the range of multi bytes character.
func randRune(runes []rune) rune {
	npair := len(runes) / 2
	rand.Seed(time.Now().UnixNano())
	i := rand.Intn(npair)
	min := int(runes[2*i])
	max := int(runes[2*i+1])

	if min == max {
		return rune(min)
	}
	randi := min + rand.Intn(max-min)
	return rune(randi)
}
