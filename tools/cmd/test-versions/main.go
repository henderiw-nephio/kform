package main

import (
	"fmt"
	"os"

	"github.com/apparentlymart/go-versions/versions"
)

func main() {
	available := versions.List{
		versions.MustParseVersion("0.0.1"),
		/*
		versions.MustParseVersion("0.8.0"),
		versions.MustParseVersion("0.0.2"),
		versions.MustParseVersion("0.0.7"),
		versions.MustParseVersion("1.0.1"),
		versions.MustParseVersion("0.9.1"),
		versions.MustParseVersion("2.0.0-beta.1"),
		versions.MustParseVersion("2.1.0"),
		versions.MustParseVersion("1.0.0"),
		versions.MustParseVersion("0.9.0"),
		versions.MustParseVersion("1.1.0"),
		versions.MustParseVersion("2.0.0"),
		*/
	}
	fmt.Println("available versions", available)
	constraints := "0.0.1"
	fmt.Println("constraints", constraints)
	allowed, err := versions.MeetingConstraintsStringRuby(constraints)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid version constraint: %s", err)
		os.Exit(1)
	}
	fmt.Println("allowed versions", allowed)

	/*
	allowed, err := versions.MeetingConstraintsStringRuby("~> 0.0.1")
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid version constraint: %s", err)
		os.Exit(1)
	}
	*/

	candidates := available.Filter(allowed)
	fmt.Println("candidate versions", candidates)
	chosen := candidates.Newest()

	hasVersion := "0.0.1"
	has := candidates.Set().Has(versions.MustParseVersion(hasVersion))
	fmt.Printf("Would install v%s\n", chosen)
	fmt.Printf("has v%s: %t\n", hasVersion, has)
}
