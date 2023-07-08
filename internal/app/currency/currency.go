package currency

import "errors"

type Currency struct {
	v string
}

var (
	UNKNOWN = Currency{"UNKNOWN"}
	RUB     = Currency{"RUB"}
	GEL     = Currency{"GEL"}
	AMD     = Currency{"AMD"}
	USD     = Currency{"USD"}
	EUR     = Currency{"EUR"}
	RSD     = Currency{"RSD"}
)

func (c Currency) String() string {
	return c.v
}

func FromString(value string) (Currency, error) {
	switch value {
	case RUB.String():
		return RUB, nil
	case GEL.String():
		return GEL, nil
	case AMD.String():
		return AMD, nil
	case USD.String():
		return USD, nil
	case EUR.String():
		return EUR, nil
	case RSD.String():
		return RSD, nil
	}

	return UNKNOWN, errors.New("Invalid currency value")
}
