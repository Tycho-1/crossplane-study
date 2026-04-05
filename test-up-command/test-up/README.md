# test-up

An example Upbound control plane project for Amazon Web Services (AWS).

A control plane project is a source-level representation of a Crossplane control
plane. It lets you treat your control plane configuration as a software project.
With a control plane project you can build your compositions using a language
like KCL or Python. This enables Crossplane schema-aware syntax highlighting,
autocompletion, and linting.

Read the [control plane project documentation][proj-docs] to learn more about
control plane projects.

This project defines a new `StorageBucket` API, which is powered by AWS S3.

[proj-docs]: https://docs.upbound.io/core-concepts/projects/
