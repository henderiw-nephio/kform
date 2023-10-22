---
title: "`push`"
linkTitle: "push"
type: docs
description: >
  Push a provider/module package in OCI format.
---

<!--mdtogo:Short
    Push a provider/module package in OCI format.
-->

`push` pushes a provider/module package to a OCI compatible registry

### Synopsis

<!--mdtogo:Long-->

```
kform pkg push DIR TAG [flags]
```

#### Args

```
DIR:
  The directory of the package. Push fails if DIR does not already exist.
TAG:
  The oci registry tag
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
# Pushes a new module-hello-world package to 
$ kform pkg push ./module-hello-world pkg.kform.io/module/module-hello-world:v0.1.0
```

<!--mdtogo-->
