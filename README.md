# cint

A configuration linter powered by CUE - validate your YAML/JSON configs against CUE schemas.

## Overview

`cint` is a command-line tool that validates configuration files against [CUE](https://cuelang.org/) schemas. Named similarly to [pint](https://github.com/cloudflare/pint) (Prometheus linter), it follows the philosophy of separation of concerns: CUE handles all validation logic declaratively, while Go acts as a thin wrapper for orchestration.

## Why cint?

- **Declarative validation**: Define constraints in CUE, not Go code
- **Type-safe**: CUE's powerful type system catches errors at validation time
- **Maintainable**: Change validation rules without touching Go code
- **Clear errors**: Precise error messages with line numbers and field paths
- **CI/CD ready**: Plain text output and proper exit codes

## Installation

```bash
$ go install github.com/zinrai/cint@latest
```

## Usage

```bash
$ cint -schema=<schema.cue> -config=<config.yaml|json> [-config=<config2.yaml|json>...]
```

### Options

- `-schema`: Path to CUE schema file (required)
- `-config`: Path to config file to validate (can be specified multiple times, supports .yaml, .yml, .json)
- `-version`: Show version

### Examples

Validate a YAML file:

```bash
$ cint -schema app.cue -config service.yaml
```

Validate a JSON file:

```bash
$ cint -schema app.cue -config service.json
```

Validate multiple files (mixed formats):

```bash
$ cint -schema app.cue -config service.yaml -config config.json
```

Using shell glob expansion:

```bash
# Note: glob expansion is handled by your shell
for file in configs/*.yaml; do
  cint -schema app.cue -config "$file"
done
```

## Example

See the `example/` directory for sample CUE schemas and configuration files:

- `example/schema.cue` - Sample CUE schema with various constraint types
- `example/valid.yaml` - Configuration that passes validation
- `example/invalid.yaml` - Configuration with validation errors (for testing)

## Exit Codes

- `0`: All files are valid
- `1`: One or more files failed validation or other errors occurred

## License

This project is licensed under the [MIT License](./LICENSE).
