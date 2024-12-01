package planner

import (
	"testing"
	"time"
)

func TestParseRules(t *testing.T) {
	source := "M2.d12.h2,M3.d5.h18,worker:3,aggregator:2@w6.h12,w7.h23,worker:5"
	result, err := parseRules(source)

	if err != nil {
		t.Errorf("%s ->\n\terror: %v", source, err)
	}

	expected := []Rule{
		{
			Start: &TimeMark{
				Month: 2,
				Day:   12,
				Hour:  2,
			},
			End: &TimeMark{
				Month: 3,
				Day:   5,
				Hour:  18,
			},
			Services: map[string]int{
				"worker":     3,
				"aggregator": 2,
			},
		},
		{
			Start: &TimeMark{
				Weekday: 6,
				Hour:    12,
			},
			End: &TimeMark{
				Weekday: 7,
				Hour:    23,
			},
			Services: map[string]int{
				"worker": 5,
			},
		},
	}

	if len(result) != len(expected) {
		t.Errorf("%s ->\n\t%v, expected: %v", source, result, expected)
	}

	for i := 0; i < len(result); i++ {
		for key, resVal := range result[i].Services {
			if expVal, ok := expected[i].Services[key]; !ok || resVal != expVal {
				t.Errorf("%s ->\n\tincorrect service %s:%d", source, key, resVal)
			}
		}

		if *result[i].Start != *expected[i].Start {
			t.Errorf("%s ->\n\tincorrect start timemark %v:%v", source, *result[i].Start, *result[i].Start)
		}

		if *result[i].End != *expected[i].End {
			t.Errorf("%s ->\n\tincorrect end timemark %v:%v", source, *result[i].End, *result[i].End)
		}
	}
}

func TestInPeriodNow(t *testing.T) {
	now := time.Now()

	rule := Rule{
		Start: &TimeMark{
			Month: int(now.Month()),
			Day:   now.Day(),
		},
		End: &TimeMark{
			Month: int(now.Month()),
			Day:   now.Day(),
		},
	}

	if !rule.InPeriodNow() {
		t.Errorf("Expected: true")
	}
}
