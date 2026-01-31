# toolcache

> **DEPRECATED**: This repository has been merged into [toolops](https://github.com/jonwraymond/toolops).
>
> All caching functionality is now available in the `github.com/jonwraymond/toolops/cache` package.
>
> See [MIGRATION.md](./MIGRATION.md) for migration instructions.

---

## Migration

This package is no longer maintained. Please migrate to:

```bash
go get github.com/jonwraymond/toolops
```

Then update your imports:

```go
// Before
import "github.com/jonwraymond/toolcache"

// After
import "github.com/jonwraymond/toolops/cache"
```

For detailed migration guidance, see [MIGRATION.md](./MIGRATION.md).

## Timeline

- **v0.x.x**: Final releases in this repository
- **toolops v1.0.0+**: All caching functionality consolidated

## Support

For questions or issues, please open issues in the [toolops repository](https://github.com/jonwraymond/toolops/issues).
