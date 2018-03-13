# Bootstrap

## Developing bootstrap scripts

If you are running `kubicorn` in the though level directory of the repository set the following environmental variable to force parse the bootstrap scripts locally.

```bash
$ KUBICORN_FORCE_LOCAL_BOOTSTRAP=1 kubicorn apply mycluster -v 4

```

These are the bootstrap scripts that ship with the default `kubicorn` profiles.

Feel free to add your own, or modify these at any time.

The scripts are effectively what we use as `user data` to initialize a VM

### I need to template out one of these bootstrap scripts

No you don't. Write bash like a pro.

### I need more data in a bootstrap script what should I do?

If you really can only get it from `kubicorn` and nowhere else, you can use the `Values{}` struct to define custom key/value pairs that will be injected into your script.
This will be a code change, and is intended to be just that.