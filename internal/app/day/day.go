package day

import (
	"errors"
	"fmt"
)

type Day struct {
	v string
}

var (
	UNKNOWN = Day{"UNKNOWN"}
	MON     = Day{"MON"}
	TUE     = Day{"TUE"}
	WED     = Day{"WED"}
	THU     = Day{"THU"}
	FRI     = Day{"FRI"}
	SAT     = Day{"SAT"}
	SUN     = Day{"SUN"}
)

func (c Day) String() string {
	return c.v
}

func FromString(value string) (Day, error) {
	switch value {
	case MON.String():
		return MON, nil
	case TUE.String():
		return TUE, nil
	case WED.String():
		return WED, nil
	case THU.String():
		return THU, nil
	case FRI.String():
		return FRI, nil
	case SAT.String():
		return SAT, nil
	case SUN.String():
		return SUN, nil
	}

	return UNKNOWN, errors.New(fmt.Sprintf("Invalid day value %s", value))
}
