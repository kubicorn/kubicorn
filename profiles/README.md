# Profiles

These are starting places for Kubernetes clusters.

By design you should be able to tweak all of the fields in these structs to fit your needs.

You can use `--set` to tweak them at runtime. For instance:

```
$ kubicorn create myCluster -p aws --set SSH.Port=9000
```

You can also modify the structs directly in Go. Feel free to copy/paste the code in this package as a starting point.

You can also use the `/examples` directory to see how to use it to reconcile a cluster.

Ideally we have a large number of profiles, for a lot of different use cases. If there is one you would like to see, please add it!


