# ArduMower Relay Server CI/CD

The Concourse pipeline in `pipeline.yml` is generated using `generate.go`.
This also happens when `task gen` is invoked from the project root.

The primary template for the generated pipeline is `pipeline.yml.tpl` but there are more:
1. The final `pipeline.yml` file is generated from the templates `pipeline.yml.tpl` and `_helpers.tpl` using `pipeline.src.yml` as final values source.
2. Before the final values source `pipeline.src.yml` is used, it is generated from the template `platforms.yml.tpl` using `platforms.src.yml` as precursor values source.

The pipeline compiles executable binaries and builds Docker images for a combination of operating systems and processor architectures.
The available combinations and their capabilities are listed in the precursor values source `platforms.src.yml`.

## Development

The most comfortable way to work on the pipeline is 
to edit the `.tpl` template files 
by switching the syntax highlighting back and forth 
between `YAML` and `Go Template`, 
or to use a smarter IDE than me.

Running the command `task gen` from the project root 
runs the template generation code 
and invokes `fly set-pipeline` with the result from the templates.