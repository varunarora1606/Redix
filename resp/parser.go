package resp

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func Parse(msg string) ([]string, error) {
	if len(msg) == 0 || msg[0] != '*' {
		return nil, errors.New("invalid RESP format")
	}

	parts := strings.Split(msg, "\r\n")
	lenth, err := strconv.Atoi(parts[0][1:])
	if err != nil {
		return nil, errors.New("invalid RESP array length")
	}

	ans := make([]string, 0, lenth)
	fmt.Println(parts)

	for i := 1; i+1 < len(parts); i++ {
		if len(parts[i]) == 0 {
			continue
		}

		switch parts[i][0] {
		case '*':
		case '$':
			if parts[i][1:] == "-1" {
				ans = append(ans, "")
				continue
			}
			if i+1 >= len(parts) {
				return nil, errors.New("unexpected end of message")
			}
			ans = append(ans, parts[i+1])
			i++
		default:
			i++
		}
	}

	return ans, nil
}
