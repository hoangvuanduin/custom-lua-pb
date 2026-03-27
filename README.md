# custom-lua-pb

A `protoc` compiler plugin that generates strictly-typed [Luau](https://luau-lang.org/) modules from Protocol Buffer definitions.

## What's Inside

- **[protoc-gen-luau/](protoc-gen-luau/)** — The plugin. Generates Luau type definitions, constructors, and JSON codecs from `.proto` files. See [protoc-gen-luau/README.md](protoc-gen-luau/README.md) for architecture, usage, and testing docs.
- **docs/** — Design specs and implementation plans
- **research/** — Background research on protoc plugin design
- **specs/** — Epic/feature specifications

## Quick Start

```bash
cd protoc-gen-luau
./generate.sh <proto_dir> [output_dir]
```

See [protoc-gen-luau/README.md](protoc-gen-luau/README.md) for full details.
