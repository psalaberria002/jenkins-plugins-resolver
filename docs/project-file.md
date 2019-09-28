
< [Prev](../README.md) (*README*) | [Next](jpresolver.md) (*Resolve project dependencies*) >

___

# Project file

The project file describes the plugins that your Jenkins project depends on.

## Syntax

It supports JSON, Jsonnet and YAML syntax:

```json
{
  "dependencies": {
    "google-login": "1.4"
  }
}
```

```yaml
dependencies:
  google-login: 1.4
```

```jsonnet
local auth_deps = {
    'google-login': '1.4',
};

{
    dependencies: auth_deps,
}
```

## Schema

The project file may include the following fields:

### dependencies

It is a map of plugins and versions:

```yaml
foo: 1.0.0
bar: 1.2.3.4
```

___

< [Prev](../README.md) (*README*) | [Next](jpresolver.md) (*Resolve project dependencies*) >
