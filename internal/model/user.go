package model

import "time"

type User struct {
	Username string
	Password string
	Country  string
	Timezone time.Duration
}
