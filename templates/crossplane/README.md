# Crossplane Provider Generation

This folder includes the templates to generate AWS Crossplane Provider. Run the
following to generate:

```console
go run -tags codegen cmd/ack-generate/main.go crossplane <resource name> \
    --output <directory for provider>
```

See [Contributing New Resource Using ACK](https://github.com/crossplane-contrib/provider-aws/blob/master/CODE_GENERATION.md)
for details.
