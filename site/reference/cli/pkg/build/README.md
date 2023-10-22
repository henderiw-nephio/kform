---
title: "`build`"
linkTitle: "build"
type: docs
description: >
  Build a provider/module package in OCI format.
---

<!--mdtogo:Short
    Build a provider/module package in OCI format.
-->

`build` builds a provider/module package in OCI format

### Synopsis

<!--mdtogo:Long-->

```
kform pkg build DIR [flags]
```

#### Args

```
DIR:
  build fails if DIR does not already exist.
```

#### Flags

```
```

<!--mdtogo-->

### Examples

{{% hide %}}

<!-- @makeWorkplace @verifyExamples-->

```
# Set up workspace for the test.
TEST_HOME=$(mktemp -d)
cd $TEST_HOME
```

{{% /hide %}}

<!--mdtogo:Examples-->

<!-- @pkgInit @verifyStaleExamples-->

```shell
# Creates a new package in OCI format for module-hello-world.
$ kform pkg build ./module-hello-world
```

<!--mdtogo-->
