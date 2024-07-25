package common

import (
	"errors"
	"time"
)

type Sprint struct {
	Number    int
	StartDate time.Time
	EndDate   time.Time
}

// getSprint returns the sprint number, start date, and end date for a given date
func GetSprint(date time.Time) (Sprint, error) {
	// Check if the date is a Saturday or Sunday
	if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
		return Sprint{}, errors.New("the date cannot be a Saturday or Sunday")
	}

	// Define the start date of the first sprint
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Adjust baseDate to the first Monday of 2024
	for baseDate.Weekday() != time.Monday {
		baseDate = baseDate.AddDate(0, 0, 1)
	}

	// Calculate the difference in days between the given date and the base date
	daysDifference := int(date.Sub(baseDate).Hours() / 24)

	// Calculate the sprint number (each sprint is 14 days long)
	sprintNumber := daysDifference/14 + 1

	// Calculate the start date of the sprint (first Monday)
	sprintStartDate := baseDate.AddDate(0, 0, (sprintNumber-1)*14)

	// Calculate the end date of the sprint (second Friday)
	sprintEndDate := sprintStartDate.AddDate(0, 0, 13)
	for sprintEndDate.Weekday() != time.Friday {
		sprintEndDate = sprintEndDate.AddDate(0, 0, -1)
	}

	return Sprint{
		Number:    sprintNumber,
		StartDate: sprintStartDate,
		EndDate:   sprintEndDate,
	}, nil
}
