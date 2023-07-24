# r8im

This is a suite of assorted tools for inspecting and manipulating
docker and OCI images for machine learning workloads.

- it does *not* depend on a local docker daemon; it directly manipulates the layers as tar files
- it works with the image registry directly
- it can modify existing images without downloading layers
- it can attach weights to an existing image
- it works from inside docker

## Configuration

Many subcommands take some or all of the following options:

 - `-t`, `--token`: replicate cog token for pushing to `r8.im`. Can also be specified as `COG_TOKEN` environment variable.
 - `-r`, `--registry`: image registry to push to (by default, `r8.im`).
 - `-h`, `--help`: get help for subcommand

## affix

Add a new layer to an existing image, without changing any of the existing layers.

```
r8im affix --base <base-image> --dest <destination-image> --tar <layer-tar-file>
```

CAUTION: `affix` can result in broken images. Because you aren't
building an image using a traditional build process, there's no
guarantees that dependencies will work correctly after manipulating an
image.

## extract

Extract weights from an image.

```
r8im extract <image> [--output file]
```

If `--output` is unspecified, weights are emitted to stdout.

Image layers are detected by searching any layer whose command ends
with ` # weights` or starts with `COPY . /src`, and within those
layers looking for appropriate files in `src/weights`.

## layers

Summarize layers of an image.

```
r8im layers <image>
```

## remix

Remix layers of an existing image. Takes one model image, and extracts
weights from a second image, combining them together into a new image.

```
r8im remix --base <image-including-tag> --weights <image-including-tag> --dest <image-dest>
```

CAUTION: `remix` can result in broken images. Because you aren't
building an image using a traditional build process, there's no
guarantees that dependencies will work correctly after manipulating an
image.

## zstd

Recompress the layers of an image using zstd.

```
r8im zstd <image> <dest>
```
