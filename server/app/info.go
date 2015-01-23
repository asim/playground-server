package app

import (
	"fmt"
)

func (i *Info) String() string {
	return fmt.Sprintf("status: %s, reason: %s, message: %s", i.Status, i.Reason, i.Message)
}
