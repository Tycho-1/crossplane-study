# Composition Functions: Where the Code Lives, Other Functions, Writing Your Own

This guide answers: **where is the code for a function** (e.g. Patch & Transform), **what other functions exist**, and **how to write your own** composition function.

---

## 1. Yes, there is code for the function

When you reference a function by name (e.g. `function-patch-and-transform`), that name refers to a **Function** resource in the cluster, which points to a **package** (an OCI image). The **code** for that function lives in a **source repository**.

For **function-patch-and-transform**:

- **Package (image):** `xpkg.crossplane.io/crossplane-contrib/function-patch-and-transform:v0.10.1`
- **Source code:** [github.com/crossplane-contrib/function-patch-and-transform](https://github.com/crossplane-contrib/function-patch-and-transform)

So yes: **crossplane-contrib/function-patch-and-transform** is the repo; the image is built from that code and published as a Crossplane Function package. You can read the code, open issues, or contribute there.

---

## 2. Other functions you can use

Besides **Patch & Transform**, you can use (or combine) other composition functions in your pipeline.

| Function | Purpose | Package / repo |
|----------|---------|----------------|
| **function-patch-and-transform** | Declarative base + patches (FromCompositeFieldPath, etc.). Simple templating, no loops. | `xpkg.crossplane.io/crossplane-contrib/function-patch-and-transform` · [crossplane-contrib/function-patch-and-transform](https://github.com/crossplane-contrib/function-patch-and-transform) |
| **function-cue** | Run **CUE** scripts to produce desired resources. Supports loops, conditionals, expressions. | `xpkg.crossplane.io/crossplane-contrib/function-cue` · [crossplane-contrib/function-cue](https://github.com/crossplane-contrib/function-cue) |
| **function-auto-ready** | Automatically detect when composed resources are **ready** (e.g. Deployment available, MR Ready). | Upbound Marketplace: [function-auto-ready](https://marketplace.upbound.io/functions/crossplane-contrib/function-auto-ready/latest) · [crossplane-contrib/function-auto-ready](https://github.com/crossplane-contrib/function-auto-ready) |
| **function-go-template** | Go-style templating for desired resources. | Contributed; check [crossplane-contrib](https://github.com/crossplane-contrib) or Marketplace. |
| **Custom (Go)** | Your own logic in **Go** (loops, conditionals, API calls). | You build from [function-template-go](https://github.com/crossplane/function-template-go) + [function-sdk-go](https://github.com/crossplane/function-sdk-go). |
| **Custom (Python)** | Your own logic in **Python**. | You build from Crossplane’s Python function template (`crossplane xpkg init`). |

**Example: add CUE for dynamic resources**

If you need **loops or conditionals** (e.g. “one bucket per entry in `spec.names`”), you can:

- Use **function-cue** and write a CUE script that iterates over the XR and emits resources, or  
- Use a **custom Go/Python function** that reads the XR and returns the desired resources.

**Example: add readiness**

To have the XR only report “ready” when composed resources (e.g. Deployment, MRs) are ready, add a pipeline step with **function-auto-ready** after the step that produces those resources.

---

## 3. How to write your own function

You can write a composition function in **Go** or **Python**. Crossplane calls it with a **request** (observed XR, optional input from the Composition) and expects a **response** (desired composed resources, optional conditions/events).

### Option A: Write a function in Go

**Prerequisites:** Go 1.23+, Docker, Crossplane CLI 1.17+.

1. **Initialize from template**
   ```bash
   crossplane xpkg init my-function function-template-go -d my-function
   cd my-function
   ```
   This uses [github.com/crossplane/function-template-go](https://github.com/crossplane/function-template-go).

2. **Rename and wire**
   - In `go.mod` and `package/crossplane.yaml`, replace the template name with your function name (e.g. `my-function`).
   - Optionally run `./init.sh <function-name>` if the template provides it.

3. **Implement logic in `fn.go`**
   - The important method is **`RunFunction(ctx, req *fnv1.RunFunctionRequest) (*fnv1.RunFunctionResponse, error)`**.
   - Use **`request.GetObservedCompositeResource(req)`** to read the XR (e.g. `xr.Resource.GetString("spec.region")`, `GetStringArray("spec.names")`).
   - Use **`request.GetInput(req, &myInput)`** if your Composition passes custom input.
   - Use **`request.GetDesiredComposedResources(req)`** to get the current desired map; add or update entries (same **resource.Name** every time for stability).
   - Use **`response.SetDesiredComposedResources(rsp, desired)`** to set the desired resources in the response.
   - Use **`response.Fatal`**, **`response.ConditionTrue/False`**, **`response.Normal/Warning`** for errors and status.

4. **Test**
   - Unit tests: `go test -v -cover .`
   - Local run: `go run . --insecure --debug` (one terminal); in another: `crossplane render xr.yaml composition.yaml functions.yaml` (with your function referenced and `render.crossplane.io/runtime: Development` on the Function so the CLI talks to your local process).

5. **Build and push**
   - Build runtime image: `docker build --platform=linux/amd64 -t runtime-amd64 .`
   - Build package: `crossplane xpkg build --package-root=package --embed-runtime-image=runtime-amd64 --package-file=function.xpkg`
   - Push: `crossplane xpkg push --package-files=function.xpkg your-registry/your-function:v0.1.0`

**SDK:** [github.com/crossplane/function-sdk-go](https://github.com/crossplane/function-sdk-go) (request/response helpers, proto types).  
**Docs:** [Write a Composition Function in Go](https://docs.crossplane.io/latest/guides/write-a-composition-function-in-go/).

### Option B: Write a function in Python

**Prerequisites:** Python 3.11+, Hatch, Docker, Crossplane CLI 1.14+.

1. **Initialize from template**
   ```bash
   crossplane xpkg init my-function function-template-python -d my-function
   cd my-function
   ```
   (Use the Python function template; exact name may vary—check Crossplane docs for “Write a Composition Function in Python”.)

2. **Implement the handler**
   - You implement a function that receives the request (observed XR, input) and returns the response (desired resources, conditions).
   - The template and [Write a Composition Function in Python](https://docs.crossplane.io/latest/guides/write-a-composition-function-in-python/) describe the exact API.

3. **Test and build**
   - Test locally (e.g. with `crossplane render` and a Development runtime), then build the image and package and push to your registry.

---

## 4. Summary

| Question | Answer |
|----------|--------|
| **Where is the code for function-patch-and-transform?** | [github.com/crossplane-contrib/function-patch-and-transform](https://github.com/crossplane-contrib/function-patch-and-transform). The image is built from that repo. |
| **What other functions can I use?** | **function-cue** (CUE scripts), **function-auto-ready** (readiness), plus custom Go/Python functions. See table above and [Upbound Marketplace](https://marketplace.upbound.io) → Functions. |
| **How do I write my own?** | **Go:** `crossplane xpkg init` from `function-template-go`, implement `RunFunction` in `fn.go` using `function-sdk-go`, test with `go test` and `crossplane render`, then build/push package. **Python:** Use the Python function template and follow the “Write a Composition Function in Python” guide. |

---

## References

- [Function Patch and Transform](https://docs.crossplane.io/latest/guides/function-patch-and-transform/)
- [Write a Composition Function in Go](https://docs.crossplane.io/latest/guides/write-a-composition-function-in-go/)
- [Write a Composition Function in Python](https://docs.crossplane.io/latest/guides/write-a-composition-function-in-python/)
- [How composition functions work](https://docs.crossplane.io/latest/composition/compositions/#how-composition-functions-work)
- [function-sdk-go](https://github.com/crossplane/function-sdk-go) · [function-template-go](https://github.com/crossplane/function-template-go)
- [crossplane-contrib/function-patch-and-transform](https://github.com/crossplane-contrib/function-patch-and-transform) · [function-cue](https://github.com/crossplane-contrib/function-cue) · [function-auto-ready](https://github.com/crossplane-contrib/function-auto-ready)
- [Upbound Marketplace – Functions](https://marketplace.upbound.io) (search for “function”)
