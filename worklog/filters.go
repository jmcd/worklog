package worklog

type EntryPredicate interface {
	ShouldAccept(e Entry) bool
}

type filter struct {
	filters    []func(Entry) bool
	input      chan Entry
	inputQuit  chan bool
	output     chan Entry
	outputQuit chan bool
}

func NewFilter(input chan Entry, inputQuit chan bool, quit chan bool) (*filter, chan Entry) {
	fc := filter{
		input:      input,
		inputQuit:  inputQuit,
		output:     make(chan Entry),
		outputQuit: quit,
	}
	return &fc, fc.output
}

func (fc *filter) shouldAccept(e Entry) bool {
	for _, a := range fc.filters {
		if !a(e) {
			return false
		}
	}
	return true
}

func (fc *filter) Add(p func(Entry) bool) {
	fc.filters = append(fc.filters, p)
}

func (fc *filter) Filter() {
	go fc.filterInternal()
}

func (fc *filter) filterInternal() {

	var e Entry

loop:
	for {
		select {

		case <-fc.inputQuit:
			fc.outputQuit <- true
			break loop

		case e = <-fc.input:

			accepted := fc.shouldAccept(e)

			if accepted {
				fc.output <- e
			}

		}
	}

}
