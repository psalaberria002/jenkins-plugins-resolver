
< [Prev](jpresolver.md) (*Resolve project dependencies*) | [Next](jpdownloader.md) (*Download project dependencies*) >

___

# Lock file

The lock file describes all the project dependencies (required and transtivie dependencies).

## Syntax

It supports JSON only.

```json
{
  "plugins": [
    {
      "name": "google-login",
      "version": "1.4"
    }
  ]
}
```

## Schema

The project file may include the following fields:

### plugins

It is a list of plugins and versions:

```json
{
  "plugins": [
    {
      "name": "foo",
      "version": "1.0.0"
    },
    {
      "name": "bar",
      "version": "1.2.3.4"
    }
  ]
}
```

___

< [Prev](jpresolver.md) (*Resolve project dependencies*) | [Next](jpdownloader.md) (*Download project dependencies*) >
