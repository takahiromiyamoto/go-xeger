package xeger

import (
	"math/rand"
	"regexp/syntax"
	"time"
)

const (
	ascii_lowercase = "abcdefghijklmnopqrstuvwxyz"
	ascii_uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	ascii_letters   = ascii_lowercase + ascii_uppercase
	digits          = "0123456789"
	punctuation     = " !\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
	control         = "\t\v\f\r"
	newline         = "\n"
	printable       = digits + ascii_letters + punctuation + control + newline
	printableNotNL  = digits + ascii_letters + punctuation + control
)

func defaultSource() Source { return rand.NewSource(time.Now().UnixNano()) }

const defaultLimit = 10

// A Source represents a source of uniformly-distributed pseudo-random int64
// values in the range [0, 1<<63).  This is a subset of interface rand.Source
// to specify only the methods required by Xeger.
type Source interface {
	Int63() int64
}

type Xeger struct {
	re     *syntax.Regexp
	source Source
	limit  int
}

func NewXeger(regex string, opts ...Option) (*Xeger, error) {
	re, err := syntax.Parse(regex, syntax.Perl)
	if err != nil {
		return nil, err
	}
	x := &Xeger{re: re}

	for _, o := range opts {
		o.apply(x)
	}

	// Set defaults
	if x.source == nil {
		x.source = defaultSource()
	}
	if x.limit <= 0 {
		x.limit = defaultLimit
	}

	return x, nil
}

// Option configures Xeger using the functional options paradigm.
type Option interface {
	apply(x *Xeger)
}

type optionFunc func(x *Xeger)

func (f optionFunc) apply(x *Xeger) {
	f(x)
}

// WithSource returns an Option that configures a Xeger with the given Source
// of random numbers.
func WithSource(s Source) Option {
	return optionFunc(func(x *Xeger) {
		x.source = s
	})
}

// WithLimit returns an Option that configures a Zeger with the given limit for
// the number of repeated subexpressions to generate (i.e., for "*" and "+"
// operators in a regular expression).
func WithLimit(limit int) Option {
	return optionFunc(func(x *Xeger) {
		x.limit = limit
	})
}

// Generate returns a string that matches the regular expression with which
// Xeger was created.
func (x *Xeger) Generate() string {
	return x.generateFromRegexp(x.re)
}

// Generates strings which are matched with re.
func (x *Xeger) generateFromRegexp(re *syntax.Regexp) string {
	switch re.Op {
	case syntax.OpLiteral: // matches Runes sequence
		return string(re.Rune)

	case syntax.OpCharClass: // matches Runes interpreted as range pair list
		sum := 0
		for i := 0; i < len(re.Rune); i += 2 {
			sum += 1 + int(re.Rune[i+1]-re.Rune[i])
		}

		index := rune(x.randInt(sum))
		for i := 0; i < len(re.Rune); i += 2 {
			delta := re.Rune[i+1] - re.Rune[i]
			if index <= delta {
				return string(rune(re.Rune[i] + index))
			}
			index -= delta + 1
		}
		return ""

	case syntax.OpAnyCharNotNL: // matches any character except newline
		c := printableNotNL[x.randInt(len(printableNotNL))]
		return string([]byte{c})

	case syntax.OpAnyChar: // matches any character
		c := printable[x.randInt(len(printable))]
		return string([]byte{c})

	case syntax.OpCapture: // capturing subexpression with index Cap, optional name Name
		return x.generateFromSubexpression(re, 1)

	case syntax.OpStar: // matches Sub[0] zero or more times
		return x.generateFromSubexpression(re, x.randInt(x.limit+1))

	case syntax.OpPlus: // matches Sub[0] one or more times
		return x.generateFromSubexpression(re, x.randInt(x.limit)+1)

	case syntax.OpQuest: // matches Sub[0] zero or one times
		return x.generateFromSubexpression(re, x.randInt(2))

	case syntax.OpRepeat: // matches Sub[0] at least Min times, at most Max (Max == -1 is no limit)
		max := re.Max
		if max == -1 {
			max = x.limit
		}
		count := x.randInt(max-re.Min+1) + re.Min
		return x.generateFromSubexpression(re, count)

	case syntax.OpConcat: // matches concatenation of Subs
		return x.generateFromSubexpression(re, 1)

	case syntax.OpAlternate: // matches alternation of Subs
		i := x.randInt(len(re.Sub))
		return x.generateFromRegexp(re.Sub[i])

		/*
			// The other cases return empty string.
			case syntax.OpNoMatch: // matches no strings
			case syntax.OpEmptyMatch: // matches empty string
			case syntax.OpBeginLine: // matches empty string at beginning of line
			case syntax.OpEndLine: // matches empty string at end of line
			case syntax.OpBeginText: // matches empty string at beginning of text
			case syntax.OpEndText: // matches empty string at end of text
			case syntax.OpWordBoundary: // matches word boundary `\b`
			case syntax.OpNoWordBoundary: // matches word non-boundary `\B`
		*/
	}

	return ""
}

// Generates strings from all sub-expressions.
// If count > 1, repeat to generate.
func (x *Xeger) generateFromSubexpression(re *syntax.Regexp, count int) string {
	b := make([]byte, 0, len(re.Sub)*count)
	for i := 0; i < count; i++ {
		for _, sub := range re.Sub {
			b = append(b, x.generateFromRegexp(sub)...)
		}
	}
	return string(b)
}

// Returns a non-negative pseudo-random number in [0,n).
// n must be > 0, but int31n does not check this; the caller must ensure it.
// randInt is simpler and faster than rand.Intn(n), because xeger just
// generates strings at random.
func (x *Xeger) randInt(n int) int {
	return int(x.source.Int63() % int64(n))
}
