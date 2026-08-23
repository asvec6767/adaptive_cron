package cronparser

import (
	"time"

	"github.com/robfig/cron/v3"
)

type Parser struct {
	parser cron.Parser
}

func NewParser() *Parser {
	return &Parser{
		parser: cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
	}
}

func (p *Parser) Parse(schedule string) (cron.Schedule, error) {
	return p.parser.Parse(schedule)
}

func (p *Parser) Next(now time.Time, schedule string) (time.Time, error) {
	sched, err := p.Parse(schedule)
	if err != nil {
		return time.Time{}, err
	}
	return sched.Next(now), nil
}

func (p *Parser) ShouldRunAt(now time.Time, schedule string) (bool, error) {
	sched, err := p.Parse(schedule)
	if err != nil {
		return false, err
	}
	next := sched.Next(now.Add(-time.Minute))
	return next.Before(now) || next.Equal(now), nil
}
