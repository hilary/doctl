/*
Copyright 2018 The Doctl Authors All rights reserved.
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

package commands

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"

	"errors"

	"github.com/digitalocean/doctl"
	"github.com/digitalocean/doctl/do"

	"github.com/stretchr/testify/assert"
)

func TestAuthCommand(t *testing.T) {
	cmd := Auth()
	assert.NotNil(t, cmd)
	assertCommandNames(t, cmd, "init", "list", "switch")
}

func TestAuthInit(t *testing.T) {
	cfw := cfgFileWriter
	DoitCmd.CmdConfigConfig.V.Set(doctl.ArgAccessToken, nil)
	defer func() {
		cfgFileWriter = cfw
	}()

	retrieveUserTokenFunc := func() (string, error) {
		return "valid-token", nil
	}

	cfgFileWriter = func() (io.WriteCloser, error) { return &nopWriteCloser{Writer: ioutil.Discard}, nil }

	withTestClient(t, func(c *CmdConfig, tm *tcMocks) {
		tm.account.EXPECT().Get().Return(&do.Account{}, nil)

		assert.NoError(t, RunAuthInit(retrieveUserTokenFunc)(c))
	})
}

func TestAuthInitWithProvidedToken(t *testing.T) {
	cfw := cfgFileWriter
	DoitCmd.CmdConfigConfig.V.Set(doctl.ArgAccessToken, "valid-token")
	defer func() {
		cfgFileWriter = cfw
		DoitCmd.CmdConfigConfig.V.Set(doctl.ArgAccessToken, nil)
	}()

	retrieveUserTokenFunc := func() (string, error) {
		return "", errors.New("should not have called this")
	}

	cfgFileWriter = func() (io.WriteCloser, error) { return &nopWriteCloser{Writer: ioutil.Discard}, nil }

	withTestClient(t, func(c *CmdConfig, tm *tcMocks) {
		tm.account.EXPECT().Get().Return(&do.Account{}, nil)

		assert.NoError(t, RunAuthInit(retrieveUserTokenFunc)(c))
	})
}

func TestAuthList(t *testing.T) {
	assert.NoError(
		t,
		RunAuthList(&CmdConfig{Out: &bytes.Buffer{}}),
	)
}

func Test_displayAuthContexts(t *testing.T) {
	testCases := []struct {
		Name     string
		Out      *bytes.Buffer
		Context  string
		Contexts map[string]interface{}
		Expected string
	}{
		{
			Name:    "default context only",
			Out:     &bytes.Buffer{},
			Context: doctl.ArgDefaultContext,
			Contexts: map[string]interface{}{
				doctl.ArgDefaultContext: true,
			},
			Expected: "default (current)\n",
		},
		{
			Name:    "default context and additional context",
			Out:     &bytes.Buffer{},
			Context: doctl.ArgDefaultContext,
			Contexts: map[string]interface{}{
				doctl.ArgDefaultContext: true,
				"test":                  true,
			},
			Expected: "default (current)\ntest\n",
		},
		{
			Name:    "default context and additional context set to addditional context",
			Out:     &bytes.Buffer{},
			Context: "test",
			Contexts: map[string]interface{}{
				doctl.ArgDefaultContext: true,
				"test":                  true,
			},
			Expected: "default\ntest (current)\n",
		},
		{
			Name:    "unset context",
			Out:     &bytes.Buffer{},
			Context: "missing",
			Contexts: map[string]interface{}{
				doctl.ArgDefaultContext: true,
				"test":                  true,
			},
			Expected: "default\ntest\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			displayAuthContexts(tc.Out, tc.Context, tc.Contexts)
			assert.Equal(t, tc.Expected, tc.Out.String())
		})
	}
}

type nopWriteCloser struct {
	io.Writer
}

var _ io.WriteCloser = (*nopWriteCloser)(nil)

func (d *nopWriteCloser) Close() error {
	return nil
}
