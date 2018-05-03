# fdb-object-store

A simple object store built on top of [FoundationDB](https://www.foundationdb.org). This is only a proof of concept.

## Running

Requires the FoundationDB Go bindings, follow the instructions [here](https://github.com/apple/foundationdb/tree/master/bindings/go) to install. Also requires the [Gin web framework](https://github.com/gin-gonic/gin).

After dependencies have been installed the server can be started with `go run main.go`.

The object store is dead simple to use and only has two features, uploading files and downloading them.

File uploads are handled by making a POST file upload to `/object/path/to/file.txt`. Everything after `/object/` will be used as the file name, there is no notion of paths or directories but they can be simulated with slashes. When uploading a content type must also be specified with the POST form field `content_type`.

```
# Example upload using HTTPie
$ http -f POST localhost:8080/object/images/my_image.png content_type="image/png" file@local_file.png
```

Downloading an existing file is as simple as making a GET request to the same past as the upload. This can also be done through a browser.

```
# Example download of previously uploaded file (once again with HTTPie)
$ http -d GET localhost:8080/object/images/my_image.png
```

## License

All code is licensed under [MIT](https://github.com/Fabianlindfors/fdb-object-store/blob/master/LICENSE).