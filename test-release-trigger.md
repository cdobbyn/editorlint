# GitHub Release Event Types

According to GitHub documentation, the `release` event has these activity types:
- `published`: A release or pre-release is published
- `unpublished`: A release or pre-release is unpublished
- `created`: A draft release is created
- `edited`: A release or pre-release is edited
- `deleted`: A release is deleted
- `prereleased`: A pre-release is created
- `released`: A release is published (not a pre-release)

The issue is that the workflow uses:
```yaml
types: [published, released]
```

But when using `softprops/action-gh-release@v2`, it creates a release directly in the "published" state. 
The "released" event type is deprecated and not triggered by modern GitHub Actions.

The correct configuration should be:
```yaml
types: [published]
```
