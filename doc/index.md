# LXD image builder

`lxd-imagebuilder` is an image building tool for LXC and LXD.

Its modern design uses pre-built official images whenever available and supports a variety of modifications on the base image.
`lxd-imagebuilder` creates LXC or LXD images, or just a plain root file system, from a declarative image definition (in YAML format) that defines the source of the image, its package manager, what packages to install or remove for specific image variants, OS releases and architectures, as well as additional files to generate and arbitrary actions to execute as part of the image build process.

`lxd-imagebuilder` can be used to create custom images that can be used as the base for LXC containers or LXD instances.

`lxd-imagebuilder` is used to build the images on the [LXD image server](https://images.lxd.canonical.com/).
You can also use it to build images from ISO files that require licenses and therefore cannot be distributed.

---

## In this documentation

````{grid} 1 1 2 2
```{grid-item} [Tutorial](tutorials/use)

**Start here**: a hands-on introduction to [`lxd-imagebuilder`](/tutorials/use) and [`simplestream-maintainer`](/tutorials/imageserver) for new users
```
```{grid-item} [](howto/index)

**Step-by-step guides** covering key operations and common tasks
```
````

````{grid} 1 1 2 2
:reverse:

```{grid-item} [](reference/index)

**Technical information** - specifications, APIs, architecture for [`lxd-imagebuilder`](/reference/lxd-imagebuilder/index) and [`simplestream-maintainer`](/reference/simplestream-maintainer/index)
```

```{grid-item} Explanation (coming)

**Discussion and clarification** of key topics
```
````

---

## Project and community

`lxd-imagebuilder` is free software and released under [AGPLv3](https://www.gnu.org/licenses/agpl-3.0.en.html).
It's an open source project that warmly welcomes community projects, contributions, suggestions, fixes and constructive feedback.

- [Contribute to the project](https://github.com/canonical/lxd-imagebuilder/blob/master/CONTRIBUTING.md)  <!-- wokeignore:rule=master -->
- [Discuss on IRC](https://web.libera.chat/#lxd) (see [Getting started with IRC](https://discourse.ubuntu.com/t/getting-started-with-irc/37907) if needed)
- [Ask and answer questions on the forum](https://discourse.ubuntu.com/c/lxd/)

```{toctree}
:hidden:
:titlesonly:

self
tutorials/index
howto/index
reference/index
```
