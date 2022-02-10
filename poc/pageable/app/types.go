package app

import (
	"fmt"
	"strings"
)

type Pageable struct {
	PageNumber int
	PageSize   int
	Sorts      []Sort
}

func (p Pageable) GetOffset() int {
	return (p.PageNumber - 1) * p.PageSize
}

func (p Pageable) GetLimit() int {
	return p.PageSize
}

type SortDiection string

const (
	SortASC  = SortDiection("ASC")
	SortDESC = SortDiection("DESC")
)

type Sort struct {
	Property  string
	Direction SortDiection
}

type SortPropertyFilter func(property string) string

func ParseSorts(sortValues []string, propertyFilter SortPropertyFilter) ([]Sort, error) {
	var sorts []Sort
	for _, sortValue := range sortValues {
		sort, err := ParseSort(sortValue, propertyFilter)
		if err != nil {
			return nil, err
		}
		sorts = append(sorts, sort)
	}
	return sorts, nil
}

func ParseSort(sortValue string, propertyFilter SortPropertyFilter) (Sort, error) {
	values := strings.Split(sortValue, ",")
	if len(values) != 2 {
		return Sort{}, fmt.Errorf("invalid sort value: %s", sortValue)
	}

	prop, directionValue := propertyFilter(values[0]), values[1]
	direction := SortDiection(strings.ToUpper(directionValue))
	switch direction {
	case SortASC, SortDESC:
		return Sort{
			Property:  prop,
			Direction: direction,
		}, nil
	default:
		return Sort{}, fmt.Errorf("invalid sort direction: %s", directionValue)
	}
}
