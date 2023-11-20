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
kform pkg push REF DIR [flags]
```

#### Args

```
DIR:
  The directory of the package. Push fails if DIR does not already exist.
REF:
  The oci registry reference
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
# Pushes a OCI package from the directory ./provider-resourcebackend to ghcr.io/kformdev/provider-resourcebackend/provider-resourcebackend:v0.0.1 with 
# - registry: ghcr.io 
# - organization/owner: kformdev
# - tag: v0.0.1
$ kform pkg push ghcr.io/kformdev/provider-resourcebackend/provider-resourcebackend:v0.0.1 ./provider-resourcebackend
```

<!--mdtogo-->
