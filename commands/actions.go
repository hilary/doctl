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
	"sort"
	"strconv"
	"time"

	"github.com/digitalocean/doctl"
	"github.com/digitalocean/doctl/commands/displayers"
	"github.com/digitalocean/doctl/do"
	"github.com/spf13/cobra"
)

// Actions creates the action commands hierarchy.
func Actions() *Command {
	cmd := &Command{
		Command: &cobra.Command{
			Use:   "action",
			Short: "action commands",
			Long:  "action is used to access action commands",
		},
	}

	CmdBuilder(cmd, RunCmdActionGet, "get <action-id>", "get action", Writer,
		aliasOpt("g"), displayerType(&displayers.Action{}))

	cmdActionList := CmdBuilder(cmd, RunCmdActionList, "list", "list actions", Writer,
		aliasOpt("ls"), displayerType(&displayers.Action{}))
	AddStringFlag(cmdActionList, doctl.ArgActionResourceType, "", "", "Action resource type")
	AddStringFlag(cmdActionList, doctl.ArgActionRegion, "", "", "Action region")
	AddStringFlag(cmdActionList, doctl.ArgActionAfter, "", "", "Action completed after in RFC3339 format")
	AddStringFlag(cmdActionList, doctl.ArgActionBefore, "", "", "Action completed before in RFC3339 format")
	AddStringFlag(cmdActionList, doctl.ArgActionStatus, "", "", "Action status")
	AddStringFlag(cmdActionList, doctl.ArgActionType, "", "", "Action type")

	cmdActionWait := CmdBuilder(cmd, RunCmdActionWait, "wait <action-id>", "wait for action to complete", Writer,
		aliasOpt("w"), displayerType(&displayers.Action{}))
	AddIntFlag(cmdActionWait, doctl.ArgPollTime, "", 5, "Re-poll time in seconds")

	return cmd
}

// RunCmdActionList run action list.
func RunCmdActionList(c *CmdConfig) error {
	actions, err := c.Actions().List()
	if err != nil {
		return err
	}

	actions, err = filterActionList(c, actions)
	if err != nil {
		return err
	}

	sort.Sort(actionsByCompletedAt(actions))

	item := &displayers.Action{Actions: actions}
	return c.Display(item)
}

type actionsByCompletedAt do.Actions

func (a actionsByCompletedAt) Len() int {
	return len(a)
}
func (a actionsByCompletedAt) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a actionsByCompletedAt) Less(i, j int) bool {
	return a[i].CompletedAt.Before(a[j].CompletedAt.Time)
}

func filterActionList(c *CmdConfig, in do.Actions) (do.Actions, error) {
	resourceType, err := c.Config.GetString(c.NS, doctl.ArgActionResourceType)
	if err != nil {
		return nil, err
	}

	region, err := c.Config.GetString(c.NS, doctl.ArgActionRegion)
	if err != nil {
		return nil, err
	}

	status, err := c.Config.GetString(c.NS, doctl.ArgActionStatus)
	if err != nil {
		return nil, err
	}

	actionType, err := c.Config.GetString(c.NS, doctl.ArgActionType)
	if err != nil {
		return nil, err
	}

	var before, after time.Time
	beforeStr, err := c.Config.GetString(c.NS, doctl.ArgActionBefore)
	if err != nil {
		return nil, err
	}

	if beforeStr != "" {
		if before, err = time.Parse(time.RFC3339, beforeStr); err != nil {
			return nil, err
		}
	}

	afterStr, err := c.Config.GetString(c.NS, doctl.ArgActionAfter)
	if err != nil {
		return nil, err
	}
	if afterStr != "" {
		if after, err = time.Parse(time.RFC3339, afterStr); err != nil {
			return nil, err
		}
	}

	out := do.Actions{}

	for _, a := range in {
		match := true

		if resourceType != "" && a.ResourceType != resourceType {
			match = false
		}

		if match && region != "" && a.RegionSlug != region {
			match = false
		}

		if match && status != "" && a.Status != status {
			match = false
		}

		if match && actionType != "" && a.Type != actionType {
			match = false
		}

		if a.CompletedAt == nil {
			match = false
		}

		if match && !isZeroTime(before) && a.CompletedAt != nil && a.CompletedAt.After(before) {
			match = false
		}

		if match && !isZeroTime(after) && a.CompletedAt != nil && a.CompletedAt.Before(after) {
			match = false
		}

		if match {
			out = append(out, a)
		}
	}

	return out, nil
}

func isZeroTime(t time.Time) bool {
	z := time.Time{}
	return z.Equal(t)
}

// RunCmdActionGet runs action get.
func RunCmdActionGet(c *CmdConfig) error {
	if len(c.Args) != 1 {
		return doctl.NewMissingArgsErr(c.NS)
	}

	id, err := strconv.Atoi(c.Args[0])
	if err != nil {
		return err
	}

	as := c.Actions()
	a, err := as.Get(id)
	if err != nil {
		return err
	}

	return c.Display(&displayers.Action{Actions: do.Actions{*a}})
}

// RunCmdActionWait waits for an action to complete or error.
func RunCmdActionWait(c *CmdConfig) error {
	if len(c.Args) != 1 {
		return doctl.NewMissingArgsErr(c.NS)
	}

	id, err := strconv.Atoi(c.Args[0])
	if err != nil {
		return err
	}

	pollTime, err := c.Config.GetInt(c.NS, doctl.ArgPollTime)
	if err != nil {
		return err
	}

	a, err := actionWait(c, id, pollTime)
	if err != nil {
		return err
	}

	return c.Display(&displayers.Action{Actions: do.Actions{*a}})
}

func actionWait(c *CmdConfig, actionID, pollTime int) (*do.Action, error) {
	as := c.Actions()

	var a *do.Action
	var err error

	for {
		a, err = as.Get(actionID)
		if err != nil {
			return nil, err
		}

		if a.Status != "in-progress" {
			break
		}

		time.Sleep(time.Duration(pollTime) * time.Second)
	}

	return a, nil
}
