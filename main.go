package main

import "os"
import "fmt"
import "flag"
import "io/ioutil"
import "./worklog"
import "time"

const dateFormat = "2006-01-02"

func main() {

	verbosePtr := flag.Bool("verbose", false, "lists all matching entries")
	sumPtr := flag.Bool("sum-duration", true, "sums duration")

	beforePtr := flag.String("before", "", "includes only entries before specified date yyyy-mm-dd")
	afterPtr := flag.String("after", "", "includes only entries after specified date yyyy-mm-dd")

	ticketPtr := flag.String("ticket", "", "includes only entries of specified ticket")

	flag.Parse()

	filename := flag.Args()[0]

	buf, err := ioutil.ReadFile(filename)

	if err != nil {
		fmt.Println("error: ", err)
		os.Exit(1)
	}

	s := string(buf)
	unfilteredQuit := make(chan bool)
	unfilteredEntryChannel := worklog.Parse(s, unfilteredQuit)
	quit := make(chan bool)
	fc, filteredEntryChannel := worklog.NewFilter(unfilteredEntryChannel, unfilteredQuit, quit)

	if beforePredicate := createEntryPredicateWithDate(beforePtr, time.Time.Before); beforePredicate != nil {
		fc.Add(beforePredicate)
	}

	if afterPredicate := createEntryPredicateWithDate(afterPtr, time.Time.After); afterPredicate != nil {
		fc.Add(afterPredicate)
	}

	if *ticketPtr != "" {
		fc.Add(func(e worklog.Entry) bool {
			for _,t := range e.Tickets {
				if t == *ticketPtr {
					return true
				}
			}
			return false
		})
	}

	processors := []entryProcessor{}

	if *verbosePtr {
		processors = append(processors, &logger{})
	}

	if *sumPtr {
		processors = append(processors, &adder{})
	}

	fc.Filter()

	var e worklog.Entry

loop:
	for {
		select {

		case <-quit:
			break loop

		case e = <-filteredEntryChannel:

			for _, p := range processors {
				p.process(e)
			}
		}
	}

	for _, p := range processors {
		p.end()
	}
}

func createEntryPredicateWithDate(sPtr *string, z func(time.Time, time.Time) bool) func(worklog.Entry) bool {
	if date, err := time.Parse(dateFormat, *sPtr); err == nil {
		return func(e worklog.Entry) bool {
			return z(e.Date, date)
		}
	}
	return nil
}

// Entry processors

type entryProcessor interface {
	process(e worklog.Entry)
	end()
}

type logger struct {
}

func (l *logger) process(e worklog.Entry) {
	fmt.Println(e)
}

func (l *logger) end() {
}

type adder struct {
	Val float64
}

func (a *adder) process(e worklog.Entry) {
	a.Val += e.Duration
}

func (a *adder) end() {
	fmt.Printf("sum of duration: %v\n", a.Val)
}
