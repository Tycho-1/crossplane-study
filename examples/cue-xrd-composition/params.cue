package crossplane

// Shared parameters — override in values.cue or values-<env>.cue
#Params: {
	// XRD / API
	group:  string | *"example.org"
	kind:   string | *"XPlatform"
	plural: string | *"xplatforms"
	scope:  "Namespaced" | "Cluster" | *"Namespaced"

	// Defaults for composed resources
	defaultRegion: string | *"us-east-2"

	// Composition metadata
	compositionName: string | *"xplatform"
}

// Default instance — override in values.cue
params: #Params
