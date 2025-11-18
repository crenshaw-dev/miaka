# Release Process

Releases are automated using GoReleaser. The workflow creates **draft releases** that you must manually publish.

## Creating a Release

1. **Tag and push:**
   ```bash
   make release VERSION=v0.1.0
   ```
   
   Or manually:
   ```bash
   git tag -a v0.1.0 -m "Release v0.1.0"
   git push origin v0.1.0
   ```

2. **Monitor the build:** Check [Actions](https://github.com/crenshaw-dev/miaka/actions) (takes 2-5 minutes)

3. **Review the draft:** Go to [Releases](https://github.com/crenshaw-dev/miaka/releases)
   - Verify all binaries are present
   - Check release notes

4. **Publish:** Click "Publish release"

⚠️ **Once published, tags and assets cannot be modified or deleted** (if immutable releases are enabled).

## If Something Goes Wrong

Delete the tag before publishing the draft:
```bash
git tag -d v0.1.0
git push origin :v0.1.0
```
Then try again with the correct tag.
