# Example client-go Vendoring Via dep

[Dep](https://github.com/golang/dep) is not yet supported by client-go as of v6.0.0. It is possible to manually make a `Gopkg.toml` that overrides every single dependency of client-go if you absolutely need to use dep. But it's not really worth the work.

This folder does include a sample `Gopkg.toml` that was made by converting the the [GoDeps.json](https://github.com/kubernetes/client-go/blob/v6.0.0/Godeps/Godeps.json) to a set of overrides. Removing duplicates and targeting root package names. It was a pain to generate and would have to be manually updated for each client-go upgrade. But it is possible to do.
