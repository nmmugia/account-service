// test/fixture/account_fixture.go
package fixture

import "account-service/src/model"

var (
	ValidCreateAccount = model.CreateAccount{
		FullName:    "Test User",
		IDNumber:    "1234567890123456",
		PhoneNumber: "081234567890",
	}
)
