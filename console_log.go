package supd

import (
	"fmt"

	"github.com/gwaylib/errors"
)

func (x *CtlCommand) Println(msg string) {
	fmt.Println(msg)
}
func (x *CtlCommand) PrintErr(err error) {
	debugErr := errors.ParseError(err)
	if x.Verbose {
		fmt.Println(err.Error())
	} else {
		fmt.Println(debugErr.Code())
	}
}
