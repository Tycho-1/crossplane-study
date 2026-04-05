package crossplane

// Production overrides — use: cue export values-prod.cue params.cue xrd.cue composition.cue ...
params: {
	group:           "example.org"
	kind:            "XPlatform"
	plural:          "xplatforms"
	defaultRegion:   "eu-central-1"
	compositionName: "xplatform-prod"
}
