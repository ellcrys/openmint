// Bootstrap file for tests. Declate common resources here
package test

import (
	"github.com/ellcrys/openmint/www"
)

func init() {

	// initialize app test config
	www.App(true, true)

}