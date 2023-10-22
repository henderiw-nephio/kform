/*
Copyright 2023 Nokia.

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

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/henderiw-nephio/kform/tools/cmd/api-gen/command"
)

func main() {
	os.Exit(runMain())
}

// runMain does the initial setup to setup logging
func runMain() int {

	// init logging
	slog.SetDefault(slog.New(
		slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{},
		),
	))

	// init context
	ctx := context.Background()
	cmd := command.GetMain(ctx)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "%s \n", err.Error())
		return 1
	}
	return 0
}
