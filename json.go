package main

import (
	"bytes"
	"fmt"
	"strconv"
	"unicode"
	"unicode/utf8"
)

// Json implement a simple Json parser with output put into Value object as in-memory DOM
// style. It is not suitable for super large Json input but easier for us to write simple
// plotter code

type JsonToken int

const (
	kJsonTokenString = iota
	kJsonTokenNumber
	kJsonTokenBoolean
	kJsonTokenNull
	kJsonTokenLSqr
	kJsonTokenRSqr
	kJsonTokenLBra
	kJsonTokenRBra
	kJsonTokenComma
	kJsonTokenColon
	kJsonTokenError
	kJsonTokenEof
)

type JsonLexeme struct {
	String  string
	Boolean bool
	Number  float64
	Token   JsonToken
	Length  int
}

type JsonLexer struct {
	Source string
	Cursor int
	Line   int
	CCount int
	Lexeme JsonLexeme
}

func newJsonLexer(source string) *JsonLexer {
	ret := &JsonLexer{
		Source: source,
		Cursor: 0,
		Line:   1,
		CCount: 1,
		Lexeme: JsonLexeme{Token: kJsonTokenNull},
	}
	ret.Next()
	return ret
}

func (l *JsonLexer) error(str string) *JsonLexeme {
	l.Lexeme.String = str
	l.Lexeme.Token = kJsonTokenError
	return &l.Lexeme
}

func (l *JsonLexer) symbol(tk JsonToken, len int) *JsonLexeme {
	l.Lexeme.Token = tk
	l.Lexeme.Length = len
	l.Cursor += len
	l.CCount += len
	return &l.Lexeme
}

func (l *JsonLexer) lexString(le int) *JsonLexeme {
	sourceLen := len(l.Source)

	b := bytes.Buffer{}
	start := l.Cursor
	l.Cursor += le
	l.CCount += le

	for l.Cursor < sourceLen {
		c, len := utf8.DecodeRuneInString(l.Source[l.Cursor:])
		if c == utf8.RuneError {
			return l.error(fmt.Sprintf("cannot decode rune starting around %d,%d", l.Line, l.CCount))
		}

		if c == '\\' {
			nc, nl := utf8.DecodeRuneInString(l.Source[l.Cursor+len:])
			if nc == utf8.RuneError {
				return l.error(fmt.Sprintf("cannot decode rune starting around %d,%d", l.Line, l.CCount+len))
			}

			switch nc {
			case 'b':
				b.WriteString("\\b")
			case 'r':
				b.WriteString("\\r")
			case 'n':
				b.WriteString("\\n")
			case 'v':
				b.WriteString("\\v")
			case 't':
				b.WriteString("\\t")
			case '\\':
				b.WriteString("\\")
			case '"':
				b.WriteString("\"")
			default:
				return l.error(fmt.Sprintf("unknown escape character \\%c around %d,%d",
					nc, l.Line, l.CCount+len))
			}
			l.Cursor += len + nl
			l.CCount += len + nl

		} else if c == '"' {
			l.Cursor += len
			l.CCount += len
			l.Lexeme.Token = kJsonTokenString
			l.Lexeme.String = b.String()
			l.Lexeme.Length = (l.Cursor - start)
			return &l.Lexeme
		} else {
			// handle all escape character rune ???
			l.Cursor += len
			l.CCount += len
			b.WriteRune(c)
		}
	}

	return l.error(fmt.Sprintf("string is not properly closed , EOF around %d,%d", l.Line, l.CCount))
}

func (l *JsonLexer) lexNumber(c rune, le int) *JsonLexeme {
	sourceLen := len(l.Source)

	b := bytes.Buffer{}
	start := l.Cursor

	if c == '-' {
		l.Cursor += le
		l.CCount += le
		b.WriteRune('-')
	} else {
		if c == '+' {
			l.Cursor += le
			l.CCount += le
		}
	}

	const (
		stateWantDotOrDigitOrEnd = iota
		stateWantDigit
		stateWantDigitOrEnd
	)

	state := stateWantDotOrDigitOrEnd

	// scan a decimal or integer number from the stream
	// until we hit a character that doesn't belong to the
	// number
LOOP:
	for l.Cursor < sourceLen {
		c, len := utf8.DecodeRuneInString(l.Source[l.Cursor:])
		switch state {
		case stateWantDotOrDigitOrEnd:
			if unicode.IsDigit(c) {
				b.WriteRune(c)
			} else if c == '.' {
				b.WriteRune(c)
				state = stateWantDigit
			} else {
				break LOOP
			}
		case stateWantDigit:
			if unicode.IsDigit(c) {
				b.WriteRune(c)
				state = stateWantDigitOrEnd
			} else {
				break LOOP
			}
		default:
			if unicode.IsDigit(c) {
				b.WriteRune(c)
			} else {
				break LOOP
			}
		}

		l.Cursor += len
		l.CCount += len
	}

	if state == stateWantDigit {
		return l.error(fmt.Sprintf("expect a digit after the \".\" around %d,%d", l.Line, l.CCount))
	}

	if value, err := strconv.ParseFloat(b.String(), 64); err != nil {
		return l.error(fmt.Sprintf("cannot parse string into float64 around %d,%d due to error %v",
			l.Line,
			l.CCount,
			err))
	} else {
		l.Lexeme.Token = kJsonTokenNumber
		l.Lexeme.Number = value
		l.Lexeme.Length = (l.Cursor - start)
		return &l.Lexeme
	}
}

func (l *JsonLexer) matchKeyword(str string) bool {
	for _, x := range str {
		nc, nl := utf8.DecodeRuneInString(l.Source[l.Cursor:])
		if nc != x {
			return false
		}
		l.Cursor += nl
		l.CCount += nl
	}

	nc, _ := utf8.DecodeRuneInString(l.Source[l.Cursor:])
	return !unicode.IsLetter(nc)
}

func (l *JsonLexer) lexKeyword(c rune, len int) *JsonLexeme {
	start := l.Cursor
	l.Cursor += len
	l.CCount += len

	if c == 't' {
		if l.matchKeyword("rue") {
			l.Lexeme.Token = kJsonTokenBoolean
			l.Lexeme.Boolean = true
			l.Lexeme.Length = (l.Cursor - start)
			return &l.Lexeme
		}
	} else if c == 'f' {
		if l.matchKeyword("alse") {
			l.Lexeme.Token = kJsonTokenBoolean
			l.Lexeme.Boolean = false
			l.Lexeme.Length = (l.Cursor - start)
			return &l.Lexeme
		}
	} else if c == 'n' {
		if l.matchKeyword("ull") {
			l.Lexeme.Token = kJsonTokenNull
			l.Lexeme.Boolean = false
			l.Lexeme.Length = (l.Cursor - start)
			return &l.Lexeme
		}
	}

	return l.error(fmt.Sprintf("unknown token around %d,%d", l.Line, l.CCount))
}

func (l *JsonLexer) Next() *JsonLexeme {
	for l.Cursor < len(l.Source) {
		c, len := utf8.DecodeRuneInString(l.Source[l.Cursor:])
		if c == utf8.RuneError {
			return l.error(fmt.Sprintf("cannot decode rune starting around %d,%d", l.Line, l.CCount))
		}
		switch c {
		case '\n':
			l.Line++
			l.CCount = 1
			l.Cursor += len
			continue
		case ' ', '\t', '\r', '\v', '\b': // whtiespace
			l.CCount++
			l.Cursor += len
			continue
		case '[':
			return l.symbol(kJsonTokenLSqr, len)
		case ']':
			return l.symbol(kJsonTokenRSqr, len)
		case '{':
			return l.symbol(kJsonTokenLBra, len)
		case '}':
			return l.symbol(kJsonTokenRBra, len)
		case ',':
			return l.symbol(kJsonTokenComma, len)
		case ':':
			return l.symbol(kJsonTokenColon, len)
		case '"':
			return l.lexString(len)
		case '-', '+', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return l.lexNumber(c, len)
		default:
			return l.lexKeyword(c, len)
		}
	}
	return l.symbol(kJsonTokenEof, 0)
}

type JsonParser struct {
	Lexer  *JsonLexer
	Source string
}

func (parser *JsonParser) error(str string) error {
	return fmt.Errorf("around %d,%d, parser has error %s", parser.Lexer.Line,
		parser.Lexer.CCount,
		str)
}

func (parser *JsonParser) parseList() (Value, error) {
	if parser.Lexer.Lexeme.Token != kJsonTokenLSqr {
		panic("expect [")
	}

	cur := parser.Lexer.Next()

	if cur.Token == kJsonTokenRSqr {
		parser.Lexer.Next()
		return Value{Type: kValueTypeList, List: NewList()}, nil
	} else {
		list := NewList()
		for {
			if value, err := parser.parseValue(); err != nil {
				return NewNull(), err
			} else {
				list.Value = append(list.Value, value)
			}

			if parser.Lexer.Lexeme.Token == kJsonTokenComma {
				parser.Lexer.Next()
			} else if parser.Lexer.Lexeme.Token == kJsonTokenRSqr {
				parser.Lexer.Next()
				break
			} else {
				return NewNull(), parser.error("expect a \"]\" or \",\" in list")
			}
		}
		return Value{Type: kValueTypeList, List: list}, nil
	}
}

func (parser *JsonParser) parseObject() (Value, error) {
	if parser.Lexer.Lexeme.Token != kJsonTokenLBra {
		panic("expect {")
	}

	cur := parser.Lexer.Next()
	if cur.Token == kJsonTokenRBra {
		parser.Lexer.Next()
		return Value{Type: kValueTypeObject, Object: NewObject()}, nil
	} else {
		obj := NewObject()
		for {
			if cur.Token != kJsonTokenString {
				return NewNull(), parser.error("expect a qutoed string as key in object")
			}

			key := cur.String

			if cur = parser.Lexer.Next(); cur.Token != kJsonTokenColon {
				return NewNull(), parser.error("expect a \":\" in object")
			}
			parser.Lexer.Next()

			if value, err := parser.parseValue(); err != nil {
				return NewNull(), err
			} else {
				obj.Value[key] = value
			}

			if parser.Lexer.Lexeme.Token == kJsonTokenComma {
				cur = parser.Lexer.Next()
			} else if parser.Lexer.Lexeme.Token == kJsonTokenRBra {
				parser.Lexer.Next()
				break
			} else {
				return NewNull(), parser.error("expect a \"}\" or \",\" in object")
			}
		}

		return Value{Type: kValueTypeObject, Object: obj}, nil
	}
}

func (parser *JsonParser) parseValue() (Value, error) {
	switch parser.Lexer.Lexeme.Token {
	case kJsonTokenNumber:
		defer parser.Lexer.Next()
		return Value{Type: kValueTypeNumber, Number: parser.Lexer.Lexeme.Number}, nil
	case kJsonTokenString:
		defer parser.Lexer.Next()
		return Value{Type: kValueTypeString, String: parser.Lexer.Lexeme.String}, nil
	case kJsonTokenBoolean:
		defer parser.Lexer.Next()
		return Value{Type: kValueTypeBoolean, Boolean: parser.Lexer.Lexeme.Boolean}, nil
	case kJsonTokenNull:
		defer parser.Lexer.Next()
		return NewNull(), nil
	case kJsonTokenLSqr:
		return parser.parseList()
	case kJsonTokenLBra:
		return parser.parseObject()
	default:
		return NewNull(), parser.error("need a number/string/null/list/object here but get something unexpected")
	}
}

func (parser *JsonParser) Parse() (Value, error) {
	var v Value
	var e error
	if parser.Lexer.Lexeme.Token == kJsonTokenLSqr {
		v, e = parser.parseList()
	} else if parser.Lexer.Lexeme.Token == kJsonTokenLBra {
		v, e = parser.parseObject()
	} else {
		return NewNull(), parser.error("expect a list/object as root")
	}

	if e != nil {
		return NewNull(), e
	}

	if parser.Lexer.Lexeme.Token != kJsonTokenEof {
		return NewNull(), parser.error("unknown text shows up after a list/object at root of json")
	}

	return v, nil
}

func NewJsonParser(source string) *JsonParser {
	ret := &JsonParser{
		Lexer:  newJsonLexer(source),
		Source: source,
	}
	return ret
}
