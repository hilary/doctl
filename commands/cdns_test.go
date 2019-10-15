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
	"testing"
	"time"

	"github.com/digitalocean/doctl"
	"github.com/digitalocean/doctl/do"
	"github.com/digitalocean/godo"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var (
	cdnID     = uuid.New().String()
	cdnOrigin = "my-spaces.nyc3.digitaloceanspaces.com"

	testCDN = do.CDN{
		CDN: &godo.CDN{
			ID:        cdnID,
			Origin:    cdnOrigin,
			Endpoint:  "my-spaces.nyc3.cdn.digitaloceanspaces.com",
			TTL:       3600,
			CreatedAt: time.Now(),
		},
	}

	testCDNWithCustomDomain = do.CDN{
		CDN: &godo.CDN{
			ID:            cdnID,
			Origin:        cdnOrigin,
			Endpoint:      "my-spaces.nyc3.cdn.digitaloceanspaces.com",
			TTL:           3600,
			CustomDomain:  "assets.myacmecorp.com",
			CertificateID: uuid.New().String(),
			CreatedAt:     time.Now(),
		},
	}

	updatedCDN = do.CDN{
		CDN: &godo.CDN{
			ID:        cdnID,
			Origin:    cdnOrigin,
			Endpoint:  "my-spaces.nyc3.cdn.digitaloceanspaces.com",
			TTL:       60,
			CreatedAt: time.Now(),
		},
	}

	updatedCDNWithCustomDomain = do.CDN{
		CDN: &godo.CDN{
			ID:            cdnID,
			Origin:        cdnOrigin,
			Endpoint:      "my-spaces.nyc3.cdn.digitaloceanspaces.com",
			TTL:           3600,
			CustomDomain:  "assets.myacmecorp.com",
			CertificateID: uuid.New().String(),
			CreatedAt:     time.Now(),
		},
	}

	testCDNList = []do.CDN{
		testCDN,
	}
)

func TestCDNCommand(t *testing.T) {
	cmd := CDN()
	assert.NotNil(t, cmd)
	assertCommandNames(t, cmd, "create", "delete", "get", "list", "update", "flush")
}

func TestCDNsGet(t *testing.T) {
	withTestClient(t, func(c *CmdConfig, tm *tcMocks) {
		tm.cdns.EXPECT().Get(cdnID).Return(&testCDN, nil)

		c.Args = append(c.Args, cdnID)

		err := RunCDNGet(c)
		assert.NoError(t, err)
	})
}

func TestCDNsGet_RequiredArguments(t *testing.T) {
	withTestClient(t, func(c *CmdConfig, tm *tcMocks) {
		err := RunCDNGet(c)
		assert.Error(t, err)
	})
}

func TestCDNsList(t *testing.T) {
	withTestClient(t, func(c *CmdConfig, tm *tcMocks) {
		tm.cdns.EXPECT().List().Return(testCDNList, nil)

		err := RunCDNList(c)
		assert.NoError(t, err)
	})
}

func TestCDNsCreate(t *testing.T) {
	withTestClient(t, func(c *CmdConfig, tm *tcMocks) {
		cdncr := &godo.CDNCreateRequest{
			Origin: cdnOrigin,
			TTL:    3600,
		}
		tm.cdns.EXPECT().Create(cdncr).Return(&testCDN, nil)

		c.Args = append(c.Args, cdnOrigin)
		c.Config.Set(c.NS, doctl.ArgCDNTTL, 3600)

		err := RunCDNCreate(c)
		assert.NoError(t, err)
	})
}

func TestCDNsCreateCustomDomain(t *testing.T) {
	withTestClient(t, func(c *CmdConfig, tm *tcMocks) {
		cdncr := &godo.CDNCreateRequest{
			Origin:        cdnOrigin,
			TTL:           3600,
			CustomDomain:  testCDNWithCustomDomain.CustomDomain,
			CertificateID: testCDNWithCustomDomain.CertificateID,
		}
		tm.cdns.EXPECT().Create(cdncr).Return(&testCDNWithCustomDomain, nil)

		c.Args = append(c.Args, cdnOrigin)
		c.Config.Set(c.NS, doctl.ArgCDNTTL, 3600)
		c.Config.Set(c.NS, doctl.ArgCDNDomain, testCDNWithCustomDomain.CustomDomain)
		c.Config.Set(c.NS, doctl.ArgCDNCertificateID, testCDNWithCustomDomain.CertificateID)

		err := RunCDNCreate(c)
		assert.NoError(t, err)
	})
}

func TestCDNsCreateCustomDomain_NoCertIDFail(t *testing.T) {
	withTestClient(t, func(c *CmdConfig, tm *tcMocks) {
		c.Args = append(c.Args, cdnOrigin)
		c.Config.Set(c.NS, doctl.ArgCDNTTL, 3600)
		c.Config.Set(c.NS, doctl.ArgCDNDomain, updatedCDNWithCustomDomain.CustomDomain)

		err := RunCDNCreate(c)
		assert.Error(t, err)
	})
}

func TestCDNsCreate_RequiredArguments(t *testing.T) {
	withTestClient(t, func(c *CmdConfig, tm *tcMocks) {
		err := RunCDNCreate(c)
		assert.Error(t, err)
	})
}

func TestCDNsCreate_ZeroFail(t *testing.T) {
	withTestClient(t, func(c *CmdConfig, tm *tcMocks) {
		c.Args = append(c.Args, cdnOrigin)
		c.Config.Set(c.NS, doctl.ArgCDNTTL, 0)

		err := RunCDNCreate(c)
		assert.Error(t, err)
	})
}

func TestCDNsUpdateTTL(t *testing.T) {
	withTestClient(t, func(c *CmdConfig, tm *tcMocks) {
		cdnur := &godo.CDNUpdateTTLRequest{
			TTL: 60,
		}
		tm.cdns.EXPECT().UpdateTTL(cdnID, cdnur).Return(&updatedCDN, nil)

		c.Args = append(c.Args, cdnID)
		c.Config.Set(c.NS, doctl.ArgCDNTTL, 60)

		err := RunCDNUpdate(c)
		assert.NoError(t, err)
	})
}

func TestCDNsUpdateTTL_ZeroFail(t *testing.T) {
	withTestClient(t, func(c *CmdConfig, tm *tcMocks) {
		c.Args = append(c.Args, cdnID)
		c.Config.Set(c.NS, doctl.ArgCDNTTL, 0)

		err := RunCDNUpdate(c)
		assert.Error(t, err)
	})
}

func TestCDNsUpdateCustomDomain(t *testing.T) {
	withTestClient(t, func(c *CmdConfig, tm *tcMocks) {
		cdnur := &godo.CDNUpdateCustomDomainRequest{
			CustomDomain:  updatedCDNWithCustomDomain.CustomDomain,
			CertificateID: updatedCDNWithCustomDomain.CertificateID,
		}
		tm.cdns.EXPECT().UpdateCustomDomain(cdnID, cdnur).Return(&updatedCDNWithCustomDomain, nil)

		c.Args = append(c.Args, cdnID)
		c.Config.Set(c.NS, doctl.ArgCDNDomain, updatedCDNWithCustomDomain.CustomDomain)
		c.Config.Set(c.NS, doctl.ArgCDNCertificateID, updatedCDNWithCustomDomain.CertificateID)

		err := RunCDNUpdate(c)
		assert.NoError(t, err)
	})
}

func TestCDNsUpdateCustomDomain_NoCertIDFail(t *testing.T) {
	withTestClient(t, func(c *CmdConfig, tm *tcMocks) {
		c.Args = append(c.Args, cdnID)
		c.Config.Set(c.NS, doctl.ArgCDNDomain, updatedCDNWithCustomDomain.CustomDomain)

		err := RunCDNUpdate(c)
		assert.Error(t, err)
	})
}

func TestCDNsUpdateRemoveCustomDomain(t *testing.T) {
	withTestClient(t, func(c *CmdConfig, tm *tcMocks) {
		cdnur := &godo.CDNUpdateCustomDomainRequest{
			CustomDomain: "",
		}
		tm.cdns.EXPECT().UpdateCustomDomain(cdnID, cdnur).Return(&testCDN, nil)

		c.Args = append(c.Args, cdnID)
		c.Config.Set(c.NS, doctl.ArgCDNDomain, "")

		err := RunCDNUpdate(c)
		assert.NoError(t, err)
	})
}

func TestCDNsUpdate_NothingToUpdateFail(t *testing.T) {
	withTestClient(t, func(c *CmdConfig, tm *tcMocks) {
		c.Args = append(c.Args, cdnID)
		err := RunCDNUpdate(c)
		assert.Error(t, err)
	})
}

func TestCDNsDelete(t *testing.T) {
	withTestClient(t, func(c *CmdConfig, tm *tcMocks) {
		tm.cdns.EXPECT().Delete(cdnID).Return(nil)

		c.Args = append(c.Args, cdnID)
		c.Config.Set(c.NS, doctl.ArgForce, true)

		err := RunCDNDelete(c)
		assert.NoError(t, err)
	})
}

func TestCDNsDelete_RequiredArguments(t *testing.T) {
	withTestClient(t, func(c *CmdConfig, tm *tcMocks) {
		err := RunCDNDelete(c)
		assert.Error(t, err)
	})
}

func TestCDNsFlushCache(t *testing.T) {
	withTestClient(t, func(c *CmdConfig, tm *tcMocks) {
		flushReq := &godo.CDNFlushCacheRequest{Files: []string{"*"}}
		tm.cdns.EXPECT().FlushCache(cdnID, flushReq).Return(nil)

		c.Args = append(c.Args, cdnID)
		c.Config.Set(c.NS, doctl.ArgCDNFiles, []string{"*"})

		err := RunCDNFlushCache(c)
		assert.NoError(t, err)
	})
}

func TestCDNsFlushCache_RequiredArguments(t *testing.T) {
	withTestClient(t, func(c *CmdConfig, tm *tcMocks) {
		err := RunCDNFlushCache(c)
		assert.Error(t, err)
	})
}
