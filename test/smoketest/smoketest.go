package main

import (
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"os"
)

var opts = godog.Options{
	Output: colors.Colored(os.Stdout),
	Format: "pretty",
	Paths:  []string{"features/smoketest"},
}

func init() {
	godog.BindCommandLineFlags("godog.", &opts)
}

func main() {
	feat := &feature{}

	suite := godog.TestSuite{
		Name:    "Smoketest",
		Options: &opts,

		ScenarioInitializer: feat.initialize,
	}

	os.Exit(suite.Run())
}
