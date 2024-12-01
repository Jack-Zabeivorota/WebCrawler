package planner

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"main/logger"
	"main/models"
	"main/msg_broker"
)

type Planner struct {
	rules  []Rule
	broker msg_broker.IMessageBroker
	log    *logger.Logger
}

// Parsers

func parseTimeMark(source string) (*TimeMark, error) {
	parts := strings.Split(source, ".")

	timemark := &TimeMark{}

	for _, part := range parts {
		if len(part) < 2 {
			return nil, errors.New("incorrect time mark")
		}

		num, err := strconv.Atoi(part[1:])

		if err != nil {
			return nil, errors.New("incorrect time mark")
		}

		if part[0] == 'M' {
			timemark.Month = num
		} else if part[0] == 'D' {
			timemark.Day = num
		} else if part[0] == 'W' {
			timemark.Weekday = num
		} else if part[0] == 'H' {
			timemark.Hour = num
		} else {
			return nil, errors.New("incorrect time mark")
		}
	}

	return timemark, nil
}

func parseServices(strServices []string) (map[string]int, error) {
	services := map[string]int{}

	for _, strService := range strServices {
		serviceAndNum := strings.Split(strService, ":")

		if len(serviceAndNum) != 2 {
			return map[string]int{}, errors.New("incorrect service")
		}

		service, strNum := serviceAndNum[0], serviceAndNum[1]

		num, err := strconv.Atoi(strNum)

		if err != nil {
			return map[string]int{}, errors.New("incorrect number of service")
		}

		services[service] = num
	}

	return services, nil
}

func parseRules(source string) ([]Rule, error) {
	// example:
	// M2.d12.h2,M3.d5.h18,worker:3,aggregator:2@w6.h12,w7.h23,worker:5

	strRules := strings.Split(source, "@")

	if strRules[0] == "" {
		return []Rule{}, nil
	}

	rules := make([]Rule, len(strRules))

	for i, strRule := range strRules {
		parts := strings.Split(strRule, ",")

		if len(parts) < 3 {
			return []Rule{}, errors.New("incorrect rule")
		}

		start, err := parseTimeMark(parts[0])

		if err != nil {
			return []Rule{}, err
		}

		end, err := parseTimeMark(parts[1])

		if err != nil {
			return []Rule{}, err
		}

		services, err := parseServices(parts[2:])

		if err != nil {
			return []Rule{}, err
		}

		rules[i] = Rule{
			Start:    start,
			End:      end,
			Services: services,
		}
	}

	return rules, nil
}

// Planner

func New(broker msg_broker.IMessageBroker, rulesConfig string) *Planner {
	log := logger.Instance()

	rules, err := parseRules(rulesConfig)

	if err != nil {
		log.Fatal("Rules parse error: %v", err)
	}

	return &Planner{
		rules:  rules,
		broker: broker,
		log:    log,
	}
}

func avr(list []int) int {
	sum := 0

	for _, val := range list {
		sum += val
	}

	return sum / len(list)
}

func (p *Planner) Run() {
	for {
		servicesEvents := map[string][]int{}

		for _, rule := range p.rules {
			if !rule.InPeriodNow() {
				continue
			}

			for service, num := range rule.Services {
				if list, ok := servicesEvents[service]; ok {
					servicesEvents[service] = append(list, num)
				} else {
					servicesEvents[service] = []int{num}
				}
			}

			services := map[string]int{}

			for service, list := range servicesEvents {
				services[service] = avr(list)
			}

			p.broker.Send(models.SignalsTopic, &models.SignMsg{
				Sign:     models.Signals.ScaleUpdate,
				Services: services,
			})
		}

		time.Sleep(time.Hour)
	}
}
