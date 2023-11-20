---
title: "`pull`"
linkTitle: "pull"
type: docs
description: >
  Pull a provider/module package in OCI format.
---

<!--mdtogo:Short
    Pull a provider/module package in OCI format.
-->

`pull` pulls a provider/module package from a OCI compatible registry

### Synopsis

<!--mdtogo:Long-->

```
kform pkg pull REF DIR [flags]
```

#### Args

```
DIR:
  The directory of the package. Pull fails if DIR does not already exist.
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
# Pulls a provider OCI package to the current directory
$ kform pkg pull ghcr.io/kformdev/provider-resourcebackend/provider-resourcebackend:v0.0.1 .
```

<!--mdtogo-->
