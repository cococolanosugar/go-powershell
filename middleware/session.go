// Copyright (c) 2017 Gorillalabs. All rights reserved.

package middleware

import (
	"fmt"
	"strings"

	"github.com/bhendo/go-powershell/utils"
	"github.com/juju/errors"
)

type session struct {
	upstream Middleware
	name     string
}

func NewSession(upstream Middleware, config *SessionConfig) (Middleware, error) {
	asserted, ok := config.Credential.(credential)
	if ok {
		credentialParamValue, err := asserted.prepare(upstream)
		if err != nil {
			return nil, errors.Annotate(err, "Could not setup credentials")
		}

		config.Credential = credentialParamValue
	}

	name := "goSess" + utils.CreateRandomString(8)
	args := strings.Join(config.ToArgs(), " ")

	fmt.Printf("try to run: %v\n", fmt.Sprintf("$%s = New-PSSession %s", name, args))

	_, _, err := upstream.Execute(fmt.Sprintf("$%s = New-PSSession %s", name, args))
	if err != nil {
		return nil, errors.Annotate(err, "Could not create new PSSession")
	}

	return &session{upstream, name}, nil
}

func (s *session) Execute(cmd string) (string, string, error) {
	fmt.Printf("try to run: %v\n", fmt.Sprintf("Invoke-Command -Session $%s -Script {%s}", s.name, cmd))

	return s.upstream.Execute(fmt.Sprintf("Invoke-Command -Session $%s -Script {%s}", s.name, cmd))
}

func (s *session) Exit() {
	fmt.Printf("try to run: %v\n", fmt.Sprintf("Disconnect-PSSession -Session $%s", s.name))

	s.upstream.Execute(fmt.Sprintf("Disconnect-PSSession -Session $%s", s.name))
	s.upstream.Exit()
}
