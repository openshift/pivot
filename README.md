pivot ➰
========
[![Build Status](https://travis-ci.org/ashcrow/pivot.svg)](https://travis-ci.org/ashcrow/pivot/)

`pivot` provides a simple command allowing you to upgrade an
OSTree-based system from an OSTree repo embedded within a container
image.

It's not intended to be run manually, but rather as part of the
installation and upgrade process of a cluster. Though one can certainly
test it today by provisioning an RHCOS node and running it directly.


Building
--------

```
make build
```

OR

```
make static
```

Example Usage
-------------

```
pivot -r $REGISTRY/os:latest
```

Though normally, one wants to use digests rather than tags, e.g.:

```
pivot -r $REGISTRY/os@sha256:fdf70521df4ed1dc135d81fd3c4608574aeca45dc22d1b4e38d16630e9d6f1a7
```

See
---
- [openshift/os](https://github.com/openshift/os/)
- [openshift/installer](https://github.com/openshift/installer)
- [openshift/machine-config-operator](https://github.com/openshift/machine-config-operator)
