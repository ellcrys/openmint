// This controller defines policies that govern
// how a request is treated before entering any main controller method
package lib

import (
	"github.com/ellcrys/openmint/extend"
)

type PolicyController struct{
	ac *AppController
}

// Create a new controller instance
func NewPolicyController(ac *AppController) *PolicyController {
	return &PolicyController{ ac }
}

// Authenticate policy.
func (pc *PolicyController) Authenticate(c *extend.Context) error {
	return nil
}