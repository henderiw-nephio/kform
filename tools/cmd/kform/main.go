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

//go:generate $GOBIN/mdtogo ./../../../site/reference/cli/init ./../../../internal/docs/generated/initdocs --license=none --recursive=true --strategy=cmdDocs
//go:generate $GOBIN/mdtogo ./../../../site/reference/cli/apply ./../../../internal/docs/generated/applydocs --license=none --recursive=true --strategy=cmdDocs
//go:generate $GOBIN/mdtogo ./../../../site/reference/cli/pkg ./../../../internal/docs/generated/pkgdocs --license=none --recursive=true --strategy=cmdDocs
//go:generate $GOBIN/mdtogo ./../../../site/reference/cli/README.md ./../../../internal/docs/generated/overview --license=none --strategy=cmdDocs

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/henderiw-nephio/kform/tools/cmd/kform/commands"
	"github.com/henderiw/logger/log"
)

const (
	defaultConfigFileSubDir = "kform"
	defaultConfigFileName   = "kform.yaml"
)

func main() {
	os.Exit(runMain())
}

// runMain does the initial setup to setup logging
func runMain() int {
	// init logging
	l := log.NewLogger(&log.HandlerOptions{Name: "kform-logger", AddSource: false})
	slog.SetDefault(l)

	// init context
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	ctx = log.IntoContext(ctx, l)

	// init cmd context
	cmd := commands.GetMain(ctx)

	if err := cmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "%s \n", err.Error())
		cancel()
		return 1
	}
	return 0
}
