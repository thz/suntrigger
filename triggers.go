// Copyright 2019 Tobias Hintze
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// github.com/thz/suntrigger
// A tool to trigger actions based on the sun's position.

package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type trigger struct {
	Kind    string // A(zimuth), S(et), R(ise)
	Degrees float64
	Action  string
}

var (
	triggers []trigger
)

func (t trigger) String() string {
	return fmt.Sprintf("At %.6f deg (%s): %q", t.Degrees, t.Kind, t.Action)
}

func ParseTrigger(s string) (t trigger) {
	parts := strings.SplitN(s, ":", 3)
	if len(parts) != 3 {
		return
	}
	f, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return
	}
	t.Degrees = f
	t.Action = parts[2]
	if parts[0] != "R" && parts[0] != "S" && parts[0] != "A" {
		return
	}
	t.Kind = parts[0]
	return
}

func (t trigger) execute() error {
	cmd := exec.Command("sh", "-c", t.Action)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	fmt.Printf("%sEOF\n", stdoutStderr)
	return nil
}
