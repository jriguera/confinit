// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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

package confinit

import (
	"fmt"
	"os"

	cli "confinit/internal/program"
	cobra "github.com/spf13/cobra"
)

var (
	program *cli.Program
	// Version is injected at compile time (from main.go)
	Version string
	// Build is injected at compile time (from main.go)
	Build string
	// Cmd represents the base command when called without any subcommands
	Cmd = &cobra.Command{
		Short:         "Applies the actions defined in the configuraion",
		Long:          `This program processes actions on folders defined in process, run the list of actions defined`,
		RunE:          run,
		SilenceUsage:  true,
		SilenceErrors: true,
		Hidden:        false,
	}
)

func init() {
	cobra.OnInitialize(initialize)
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	program = cli.NewProgram(Build, Version, "config", Cmd)
}

// Run adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the Cmd.
func Run(version, build string) {
	Version = version
	Build = build
	if err := Cmd.Execute(); err != nil {
		fmt.Printf("Errors:\n")
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
}

// Main Callback
func run(command *cobra.Command, args []string) error {
	err := program.LoadConfig()
	if err == nil {
		return program.RunAll()
	}
	return err
}

// initialize sets up the program
func initialize() {
	program.Init()
}
