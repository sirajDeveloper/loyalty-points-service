package service

type OrderNumberValidator interface {
	Validate(number string) bool
}

type luhnOrderNumberValidator struct{}

func NewLuhnOrderNumberValidator() OrderNumberValidator {
	return &luhnOrderNumberValidator{}
}

func (v *luhnOrderNumberValidator) Validate(number string) bool {
	if len(number) < 2 {
		return false
	}

	sum := 0
	alternate := false

	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')
		if digit < 0 || digit > 9 {
			return false
		}

		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}
