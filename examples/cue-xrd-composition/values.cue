package crossplane

// Override params here. CUE merges this with params.cue.
// Edit this file to change group, kind, region, etc.
params: {
	group:          "example.org"
	kind:           "XPlatform"
	plural:         "xplatforms"
	defaultRegion:  "us-east-2"
	compositionName: "xplatform"
}
