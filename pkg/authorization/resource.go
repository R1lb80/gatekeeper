/*
Copyright 2015 All rights reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package authorization

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gogatekeeper/gatekeeper/pkg/constant"
	"github.com/gogatekeeper/gatekeeper/pkg/utils"
)

// Resource represents a url resource to protect
type Resource struct {
	// URL the url for the resource
	URL string `json:"uri" yaml:"uri"`
	// Methods the method type
	Methods []string `json:"methods" yaml:"methods"`
	// WhiteListed permits the prefix through
	WhiteListed bool `json:"white-listed" yaml:"white-listed"`
	// RequireAnyRole indicates that ANY of the roles are required, the default is all
	RequireAnyRole bool `json:"require-any-role" yaml:"require-any-role"`
	// Headers required to access this url
	Headers []string `json:"headers" yaml:"headers"`
	// Roles the roles required to access this url
	Roles []string `json:"roles" yaml:"roles"`
	// Groups is a list of groups the user is in
	Groups []string `json:"groups" yaml:"groups"`
	// AcrValues is a list of acr to add in the auth url
	AcrValues []string `json:"acr-values" yaml:"acr-values"`
}

func NewResource() *Resource {
	return &Resource{
		Methods: utils.AllHTTPMethods,
	}
}

/*
	parse decodes a resource definition
*/
//nolint:cyclop
func (r *Resource) Parse(resource string) (*Resource, error) {
	if resource == "" {
		return nil, errors.New("the resource has no options")
	}

	for _, x := range strings.Split(resource, "|") {
		keyPair := strings.Split(x, "=")

		keyPairMembers := 2
		if len(keyPair) != keyPairMembers {
			return nil,
				errors.New(
					"invalid resource keypair, should be " +
						"(uri|roles|headers|methods|white-listed|acr-values)=comma_values",
				)
		}

		switch keyPair[0] {
		case "uri":
			r.URL = keyPair[1]

			if !strings.HasPrefix(r.URL, "/") {
				return nil, errors.New("the resource uri should start with a '/'")
			}
		case "methods":
			r.Methods = strings.Split(keyPair[1], ",")

			if len(r.Methods) == 1 {
				if strings.EqualFold(r.Methods[0], constant.AnyMethod) {
					r.Methods = utils.AllHTTPMethods
				}
			}
		case "acr-values":
			r.AcrValues = strings.Split(keyPair[1], ",")
		case "require-any-role":
			val, err := strconv.ParseBool(keyPair[1])

			if err != nil {
				return nil, err
			}

			r.RequireAnyRole = val
		case "roles":
			r.Roles = strings.Split(keyPair[1], ",")
		case "headers":
			r.Headers = strings.Split(strings.ToLower(keyPair[1]), ",")
			colonCount := strings.Count(keyPair[1], ":")
			if len(r.Headers) != colonCount {
				return nil, errors.New("headers key and value should be split by colon")
			}
		case "groups":
			r.Groups = strings.Split(keyPair[1], ",")
		case "white-listed":
			value, err := strconv.ParseBool(keyPair[1])

			if err != nil {
				return nil, errors.New(
					"the value of whitelisted must be " +
						"true|TRUE|T or it's false equivalent",
				)
			}

			r.WhiteListed = value
		default:
			return nil,
				errors.New("invalid identifier, should be roles, uri or methods")
		}
	}

	return r, nil
}

// valid ensure the resource is valid
func (r *Resource) Valid() error {
	if r.Methods == nil {
		r.Methods = make([]string, 0)
	}

	if r.Roles == nil {
		r.Roles = make([]string, 0)
	}

	if r.URL == "" {
		return errors.New("resource does not have url")
	}

	if strings.HasSuffix(r.URL, "/") && !r.WhiteListed {
		if r.URL != "/" {
			return fmt.Errorf(
				"you need a wildcard on the url resource "+
					"to cover all request i.e. --resources=uri=%s*",
				r.URL,
			)
		}
	}

	// step: add any of no methods
	if len(r.Methods) == 0 {
		r.Methods = utils.AllHTTPMethods
	}
	// step: check the method is valid
	for _, m := range r.Methods {
		if !utils.IsValidHTTPMethod(m) {
			return fmt.Errorf("invalid method %s", m)
		}
	}

	return nil
}

// GetRoles returns a list of roles for this resource
func (r Resource) GetRoles() string {
	return strings.Join(r.Roles, ",")
}
// GetAcrValues returns a list of Acr for this resource
func (r Resource) GetAcrValues() string {
	return strings.Join(r.AcrValues, ",")
}

// GetHeaders returns a list of headers for this resource
func (r Resource) GetHeaders() string {
	return strings.Join(r.Headers, ",")
}

// String returns a string representation of the resource
func (r Resource) String() string {
	if r.WhiteListed {
		return fmt.Sprintf("uri: %s, white-listed", r.URL)
	}
    acrs :=""
	roles := "authentication only"
	methods := constant.AnyMethod

	if len(r.Roles) > 0 {
		roles = strings.Join(r.Roles, ",")
	}
	if len(r.AcrValues) > 0 {
		acrs = strings.Join(r.AcrValues, ",")
	}

	if len(r.Methods) > 0 {
		methods = strings.Join(r.Methods, ",")
	}

	return fmt.Sprintf("uri: %s, methods: %s, required: %s,with acr :%s", r.URL, methods, roles,acrs)
}
