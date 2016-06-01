package base62

import (
	"errors"
)

// characters used for conversion
const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// converts number to base62
func Encode(number int64) string {
	if number == 0 {
		return string(alphabet[0])
	}

	chars := make([]byte, 0)

	length := int64(len(alphabet))

	for number > 0 {
		result := number / length
		remainder := number % length
		chars = append(chars, alphabet[remainder])
		number = result
	}

	for i, j := 0, len(chars)-1; i < j; i, j = i+1, j-1 {
		chars[i], chars[j] = chars[j], chars[i]
	}

	return string(chars)
}

// converts base62 token to int
func Decode(token string) (int64, error) {
	var n uint64 = 0

	for i := 0; i < len(token); i++ {
		var v byte
		d := token[i]
		switch {
		case '0' <= d && d <= '9':
			v = d - '0'
		case 'a' <= d && d <= 'z':
			v = d - 'a' + 10
		case 'A' <= d && d <= 'Z':
			v = d - 'A' + 36
		default:
			return 0, errors.New("Invalid token string to decode")
		}
		n *= 62
		n += uint64(v)
	}

	return int64(n), nil
}
