# split-layer-psd

Extracts each layers in psd file, and output them to image files.

![howtouse](howtouse.jpg)

This tool is based on `github.com/oov/psd`.
 
## Install

~~~
go install github.com/egawata/split-layer-psd@latest
~~~

## Usage

~~~
split-layer-psd -f image.psd -o outdir
~~~

## Options

- `-f`: (*required*) psd filename
- `-o`: directory to save images. If omitted, the same directory with psd file is used.
