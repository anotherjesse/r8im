# r8im

This is a suite of assorted tools for inspecting and manipulating docker and OCI images.

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

## zstd

Recompress the layers of an image using zstd.

Examples
--------

Here are some examples of how to use the tool:

* To report the layers of an image, run:
```
$ r8im report my-image
```
This will output the layers of the specified image, along with their sizes and hashes.

* To recompress an image using zstd, run:
```
$ r8im recompress my-image
```
This will recompress the specified image using zstd, which can reduce the size of the image.

* To add a new layer to an image, run:
```
$ r8im add-layer my-image layer
```


- doesn't require or use docker / containers
- works with image registry directly
- attach weights (or other changes) to an existing image
- modify existing images without downloading layers
- works inside of docker

## story time: the nightmare

Imagine the perfect image. 

I'll wait

Oh... Oh no!  We need to make a change.  We need to add the new weights.

ok, I guess I need: 

- docker
- cog
- a GPU
- lots of bandwidth and time

ugh.  I just want to add this ontop of that.  And give it back to you.  Why are you making me do all this work!

ok...  :sadpanda:

## story time: a new day?

Oh, some nice programmer created a new tool for me.

It lets me talk to replicate's image registry - and tweak and image without the nightmare / crying...

Let see, how can I use it?

    $ ./r8im affix --help
    
    Usage of ./r8-affix:
      -base string
            base image reference - include tag: r8.im/username/modelname@sha256:hexdigest
      -dest string
            destination image reference: r8.im/username/modelname
      -registry string
            registry host (default "r8.im")
      -tar string
            tar file to append as new layer
      -token string
            replicate cog token

Some of those are confusing.  What is the tar file exactly?  (Perhaps that programmer isn't as nice as they seemed at first ..)

Let's see if I can use this anyway.

I have this file called `weights.tar` that I got from my custom dreambooth trainer.  And I have a "cog token" from https://replicate.com/auth/token


    ./r8im affix --token $REPLICATE_COG_TOKEN \
      --base "r8.im/replicate/dreambooth-template@sha256:d0b01c9e0d4bc94c8d642064b349261c0d5147a784dab8011c0adb77fe0b27d3 \
      --dest "r8.im/anotherjesse/my-dreambooths" \
      --tar "weights.tar"
      
    fetching metadata for r8.im/replicate/dreambooth-template@sha256:d0b01c9e0d4bc94c8d642064b349261c0d5147a784dab8011c0adb77fe0b27d3
    pulling took 325.867939ms
    appending as new layer /home/jesse/output.tar
    appending took 29.523629917s
    pushing took 18.92398947s
    r8.im/anotherjesse/faster@sha256:f5406d243df29db34ea441401141bf7f0f79da679651f110871a78d37c897c73


yay, it looks like I have a new version of `my-dreambooths` that added my weights on top of my template

- I didn't have to install docker or nvidia gpu just to add my weights
- I didn't have to wait for the layers from `dreambooth-template` to download, I just downloaded the metadata
- This can run anywhere! Perhaps it can even run INSIDE replicate!?  So I don't need to do anything?

Oh, I just noticed that if my tar file has an update for predict.py, it can result in a broken image.  Perhaps this is a tool to use with caution?
