package worklog

import "fmt"
import "strings"
import "unicode/utf8"
import "unicode"

const eof = -1
const digits = "0123456789"

type item struct {
	typ itemType
	val string
}

func (i item) String() string {
	switch i.typ {
	case itemEOF:
		return "EOF"
	case itemError:
		return i.val
	}

	var s string
	if len(i.val) > 10 {
		s = fmt.Sprintf("%.10q...", i.val)
	} else {
		s = fmt.Sprintf("%q", i.val)
	}

	s = fmt.Sprintf("%v %v", i.typ, s)

	return s
}

type itemType int

const (
	itemError itemType = iota
	itemEOF
	itemDate
	itemDuration
	itemText
	itemTicket
)

type stateFn func(*lexer) stateFn

type lexer struct {
	name  string
	input string
	start int
	pos   int
	width int
	items chan item
}

func lex(input string) chan item {
	l := &lexer{
		input: input,
		items: make(chan item),
	}
	go l.run()
	return l.items
}

func (l *lexer) run() {
	for state := lexText; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) backup() {
	l.pos -= l.width
}
func (l *lexer) peek() rune {
	rune := l.next()
	l.backup()
	return rune
}

func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptN(valid string, n int) bool {
	for i := 0; i < n; i++ {
		if !l.accept(valid) {
			return false
		}
	}
	return true
}

func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{
		itemError,
		fmt.Sprintf(format, args...),
	}
	return nil
}

func lexText(l *lexer) stateFn {
	var emitItemTextWhereExists func()
	emitItemTextWhereExists = func() {
		if l.pos > l.start {
			l.emit(itemText)
		}
	}
	for {
		if strings.HasPrefix(l.input[l.pos:], "+") {
			emitItemTextWhereExists()
			return lexDuration
		}

		if strings.HasPrefix(l.input[l.pos:], "@") {
			emitItemTextWhereExists()
			return lexDate
		}

		if strings.HasPrefix(l.input[l.pos:], "#") {
			emitItemTextWhereExists()
			return lexTicket
		}

		if l.next() == eof {
			break
		}
	}

	emitItemTextWhereExists()
	l.emit(itemEOF)
	return nil
}

func lexDuration(l *lexer) stateFn {

	l.accept("+")
	l.acceptRun(digits)
	if l.accept(".") {
		l.acceptRun(digits)
	}

	l.emit(itemDuration)
	return lexText
}

func lexDate(l *lexer) stateFn {
	l.accept("@")
	l.acceptN(digits, 4)
	l.accept("-")
	l.acceptN(digits, 2)
	l.accept("-")
	l.acceptN(digits, 2)
	l.emit(itemDate)
	return lexText
}

func lexTicket(l *lexer) stateFn {
	l.accept("#")
	l.acceptRun(digits)
	l.emit(itemTicket)
	return lexText
}

func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
