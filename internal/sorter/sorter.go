package sorter

import (
	"reflect"
	"sort"

	. "github.com/achannarasappa/ticker/internal/position"
	. "github.com/achannarasappa/ticker/internal/quote"

	"github.com/novalagung/gubrak/v2"
)

type sortFunction func(quotes []Quote, positions map[string]Position) []Quote

type Sorter struct {
	fn          sortFunction
	Reverse     bool
	Description string
}

func NewSorter(sort string) Sorter {
	s := Sorter{}
	if sorter, ok := sortDict[sort]; ok {
		s.fn = sorter
		s.Description = sort
	} else {
		s.fn = sortByChange
		s.Description = "change"
	}
	return s
}

func (s Sorter) Sort(quotes []Quote, positions map[string]Position) []Quote {
	q := s.fn(quotes, positions)
	if s.Reverse {
		q = gubrak.From(q).Reverse().Result().([]Quote)
	}
	return q
}

func (s Sorter) NextSorter() Sorter {
	keys := sortKeys()
	for i, sort := range keys {
		if reflect.ValueOf(s.fn).Pointer() == reflect.ValueOf(sortDict[sort]).Pointer() && i < len(keys)-1 {
			return NewSorter(keys[i+1])
		}
	}
	return NewSorter(keys[0])
}

var sortDict = map[string]sortFunction{
	"alpha":          sortByTicker,
	"change":         sortByChange,
	"dividend date":  sortByDividendDate,
	"dividend yield": sortByDividendYield,
	"pb":             sortByPriceToBook,
	"pe":             sortByPriceToEarnings,
	"value":          sortByValue,
}

func sortByTicker(quotes []Quote, positions map[string]Position) []Quote {
	if len(quotes) <= 0 {
		return quotes
	}

	result := gubrak.
		From(quotes).
		OrderBy(func(v Quote) string {
			return v.Symbol
		}).
		Result()

	return (result).([]Quote)
}

func sortByValue(quotes []Quote, positions map[string]Position) []Quote {
	if len(quotes) <= 0 {
		return quotes
	}

	activeQuotes, inactiveQuotes := splitActiveQuotes(quotes)

	cActiveQuotes := gubrak.From(activeQuotes)
	cInactiveQuotes := gubrak.From(inactiveQuotes)

	positionsSorter := func(v Quote) float64 {
		return positions[v.Symbol].Value
	}

	cActiveQuotes.OrderBy(positionsSorter, false)
	cInactiveQuotes.OrderBy(positionsSorter, false)

	result := cActiveQuotes.
		Concat(cInactiveQuotes.Result()).
		Result()

	return (result).([]Quote)
}

func sortByChange(quotes []Quote, positions map[string]Position) []Quote {
	if len(quotes) <= 0 {
		return quotes
	}

	activeQuotes, inactiveQuotes := splitActiveQuotes(quotes)

	cActiveQuotes := gubrak.
		From(activeQuotes)

	cActiveQuotes.OrderBy(func(v Quote) float64 {
		return v.ChangePercent
	}, false)

	result := cActiveQuotes.
		Concat(inactiveQuotes).
		Result()

	return (result).([]Quote)
}

func sortByPriceToBook(q []Quote, positions map[string]Position) []Quote {
	if len(q) <= 0 {
		return q
	}

	sorted := gubrak.
		From(q).
		OrderBy(func(v Quote) float64 {
			return v.PriceToBook
		}, false).
		Result()

	return (sorted).([]Quote)
}

func sortByPriceToEarnings(q []Quote, positions map[string]Position) []Quote {
	if len(q) <= 0 {
		return q
	}

	sorted := gubrak.
		From(q).
		OrderBy(func(v Quote) float64 {
			return v.TrailingPE
		}, false).
		Result()

	return (sorted).([]Quote)
}

func sortByDividendDate(q []Quote, positions map[string]Position) []Quote {
	if len(q) <= 0 {
		return q
	}

	hasDividendDate, noDividendDate, _ := gubrak.
		From(q).
		Partition(func(v Quote) bool {
			return !v.DividendDate.IsZero()
		}).
		ResultAndError()

	sorted := gubrak.
		From(hasDividendDate).
		OrderBy(func(v Quote) int64 {
			return v.DividendDate.Unix()
		}, true)

	result := sorted.
		Concat(noDividendDate).
		Result()

	return (result).([]Quote)
}

func sortByDividendYield(q []Quote, positions map[string]Position) []Quote {
	if len(q) <= 0 {
		return q
	}

	sorted := gubrak.
		From(q).
		OrderBy(func(v Quote) float64 {
			return v.DividendYield
		}, false).
		Result()

	return (sorted).([]Quote)
}

func splitActiveQuotes(quotes []Quote) (interface{}, interface{}) {
	activeQuotes, inactiveQuotes, _ := gubrak.
		From(quotes).
		Partition(func(v Quote) bool {
			return v.IsActive
		}).
		ResultAndError()

	return activeQuotes, inactiveQuotes
}

func sortKeys() []string {
	keys := make([]string, len(sortDict))
	i := 0
	for k := range sortDict {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}
