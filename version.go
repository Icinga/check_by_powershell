package main

// default vars set by goreleaser
// https://goreleaser.com/customization/build/
var (
	version = "0.2.0"
	commit  string
	date    string
	builtBy string
)

func buildVersion() string {
	s := version

	if commit != "" {
		s += " - " + commit
	}

	if date != "" {
		s += " (" + date + ")"
	}

	if builtBy != "" {
		s += " (built by " + builtBy + ")"
	}

	return s
}
