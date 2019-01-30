pivot ➰
========
[![Build Status](https://travis-ci.org/openshift/pivot.svg)](https://travis-ci.org/openshift/pivot/)

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

It also comes with a systemd unit to provide a "host API". For example:

```
mkdir -p /etc/pivot
echo $REGISTRY/os:latest > /etc/pivot/image-pullspec
touch /run/pivot-reboot-needed
systemctl start pivot
```

This will start `pivot`, which will read the file and execute the pivot.
If the pivot is completed, the file will be deleted. The expected way to
make use of this is to create the necessary files from Ignition.

See
---
- [openshift/os](https://github.com/openshift/os/)
- [openshift/installer](https://github.com/openshift/installer)
- [openshift/machine-config-operator](https://github.com/openshift/machine-config-operator)
