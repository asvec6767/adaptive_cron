package utils

import (
    "time"
    "github.com/robfig/cron/v3"
)

func ShouldRunAt(now time.Time, schedule string) (bool, error) {
    parser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
    
    sched, err := parser.Parse(schedule)
    if err != nil {
        return false, err
    }
    
    next := sched.Next(now.Add(-time.Minute))
    return next.Before(now) || next.Equal(now), nil
}

func GetNextRunTime(now time.Time, schedule string) (time.Time, error) {
    parser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
    
    sched, err := parser.Parse(schedule)
    if err != nil {
        return time.Time{}, err
    }
    
    return sched.Next(now), nil
}

func CalculateBackoff(attempt int, baseDelay time.Duration) time.Duration {
    // Экспоненциальная задержка: baseDelay * 2^attempt
    delay := baseDelay * time.Duration(1<<uint(attempt))
    maxDelay := 1 * time.Hour
    if delay > maxDelay {
        delay = maxDelay
    }
    return delay
}