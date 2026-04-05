package crossplane

// XPlatform XRD schema — uses params from values.cue
#XRD: {
	apiVersion: "apiextensions.crossplane.io/v2"
	kind:       "CompositeResourceDefinition"
	metadata: {
		name: "\(params.plural).\(params.group)"
	}
	spec: {
		scope: params.scope
		group: params.group
		names: {
			kind:  params.kind
			plural: params.plural
		}
		versions: [{
			name:          "v1alpha1"
			served:        true
			referenceable: true
			schema: {
				openAPIV3Schema: {
					type: "object"
					properties: {
						spec: {
							type: "object"
							properties: {
								region: {
									type:    "string"
									default: params.defaultRegion
								}
								storageName:   { type: "string" }
								tableName:     { type: "string" }
								iamPolicyName: { type: "string" }
							}
							required: ["region", "storageName", "tableName", "iamPolicyName"]
						}
					}
				}
			}
		}]
	}
}

xrd: #XRD
