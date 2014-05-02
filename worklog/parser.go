package worklog

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const dateFormat = "2006-01-02"

type Entry struct {
	Body string
	Date time.Time
	Duration float64
	Tickets []string
}

func (e Entry) String() string {
	return fmt.Sprintf("%v %v %v %v", e.Date.Format(dateFormat), e.Duration, e.Body, e.Tickets)
}

type parser struct {
	entries chan Entry
	quit chan bool
}

func Parse(input string, quit chan bool) chan Entry {
	p := &parser{
		entries: make(chan Entry),
		quit: quit,
	}

	items := lex(input)

	go p.run(items)
	return p.entries
}

func (p *parser) emit(body string, t time.Time, duration float64, tickets []string) {

	if !t.IsZero() {
		e := Entry{body, t, duration, tickets}
		p.entries <- e
	}
}

func (p *parser) end() {
	p.quit <- true
}

func (p *parser) run(items chan item) {

	body := ""
	var date time.Time
	duration := 0.0
	var tickets []string

	for i := range items {

		// This is rubbish, does not scale
		switch i.typ {
		case itemEOF:
			p.emit(body, date, duration, tickets)
			p.end()
		case itemDate:

			p.emit(body, date, duration, tickets)

			d, e := time.Parse(dateFormat, i.val[1:])
			if e != nil {
			}

			date = d
			duration = 0.0
			body = ""
			tickets = make([]string, 0)

		case itemDuration:
			d, e := strconv.ParseFloat(i.val, 64)
			if e == nil {
				duration += d
			}

		case itemTicket:
			t := i.val[1:]
			tickets = append(tickets, t)

		case itemText:
			trimmedVal := strings.Trim(i.val, " \n\t")
			if body != "" {
				body += " " + trimmedVal

			} else {
				body = trimmedVal
			}
		}
	}
}
