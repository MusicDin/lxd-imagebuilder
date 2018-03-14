# distrobuilder
System container image builder for LXC and LXD

## Example yaml file

Save the following yaml as a file (for example `ubuntu.yaml`). To create
a simple `Ubuntu` rootfs in a folder called `ubuntu-rootfs` call
`distrobuilder` as `distrobuilder build-dir ubuntu.yaml ubuntu-rootfs`.

```yaml
image:
  distribution: ubuntu # required
  release: xenial # optional
  variant: default # optional
  description: Ubuntu Artful # optional
  expiry: 30d # optional: defaults to 30d
  architecture: x86_64 # optional: defaults to local architecture

source:
  downloader: debootstrap
  url: http://us.archive.ubuntu.com/ubuntu
  keys:
    - 0x790BC7277767219C42C86F933B4FE6ACC0B21F3v
  keyserver: pgp.mit.edu # optional

  apt_source:: |-
    deb {{ source.URL }} {{ image.Release }} main restricted universe multiverse
    deb {{ source.URL }} {{ image.Release }}-updates main restricted universe multiverse
    deb http://security.ubuntu.com/ubuntu {{ image.Release }}-security main restricted universe multiverse

targets:
  lxc:
    create-message: |-
        You just created an {{ image.Description }} container.

        To enable SSH, run: apt install openssh-server
        No default root or user password are set by LXC.

    config:
      - type: all
        before: 5
        content: |-
          lxc.include = LXC_TEMPLATE_CONFIG/ubuntu.common.conf

      - type: user
        before: 5
        content: |-
          lxc.include = LXC_TEMPLATE_CONFIG/ubuntu.userns.conf

      - type: all
        after: 4
        content: |-
          lxc.include = LXC_TEMPLATE_CONFIG/common.conf

      - type: user
        after: 4
        content: |-
          lxc.include = LXC_TEMPLATE_CONFIG/userns.conf

      - type: all
        content: |-
          lxc.arch = x86_64

files:
 - path: /etc/hostname
   generator: hostname

 - path: /etc/hosts
   generator: hosts

 - path: /etc/resolvconf/resolv.conf.d/original
   generator: remove

 - path: /etc/resolvconf/resolv.conf.d/tail
   generator: remove

 - path: /etc/machine-id
   generator: remove

 - path: /etc/netplan/10-lxc.yaml
   generator: dump
   content: |-
     network:
       ethernets:
         eth0: {dhcp4: true}
     version: 2
   releases:
     - artful
     - bionic

 - path: /etc/network/interfaces
   generator: dump
   content: |-
     # This file describes the network interfaces available on your system
     # and how to activate them. For more information, see interfaces(5).

     # The loopback network interface
     auto lo
     iface lo inet loopback

     auto eth0
     iface eth0 inet dhcp
   releases:
     - trusty
     - xenial

 - path: /etc/init/lxc-tty.conf
   generator: upstart-tty
   releases:
    - precise
    - trusty

packages:
    manager: apt

    update: true
    install:
        - apt-transport-https
        - language-pack-en
        - openssh-client
        - vim

actions:
    - trigger: post-update
      action: |-
        #!/bin/sh
        set -eux

        # Create the ubuntu user account
        getent group sudo >/dev/null 2>&1 || groupadd --system sudo
        useradd --create-home -s /bin/bash -G sudo -U ubuntu

    - trigger: post-packages
      action: |-
        #!/bin/sh
        set -eux

        # Make sure the locale is built and functional
        locale-gen en_US.UTF-8
        update-locale LANG=en_US.UTF-8

        # Cleanup underlying /run
        mount -o bind / /mnt
        rm -rf /mnt/run/*
        umount /mnt

        # Cleanup temporary shadow paths
        rm /etc/*-

mappings:
  architecture_map: debian
```
