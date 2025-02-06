package utils

import (
	"fmt"

	"golang.org/x/exp/rand"
)

// generateAccountNumber generates a random 10-digit account number.
func GenerateAccountNumber() string {
	min := 1000000000
	max := 9999999999
	return fmt.Sprintf("%d", rand.Intn(max-min)+min)
}
