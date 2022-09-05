# split_layer_psd

Extracts each layers in psd file, and output them to image files.

This tool is based on `github.com/oov/psd`.
 

## Usage

~~~
split_layer_psd -f image.psd -o outdir
~~~

## Options

- `-f`: (*required*) psd filename
- `-o`: directory to save images. If omitted, the same directory with psd file is used.
