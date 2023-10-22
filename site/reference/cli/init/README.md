---
title: "`init`"
linkTitle: "init"
type: docs
description: >
  Initialize a new or existing kform working directory by creating initial files, initializing backend state, downloading modules, providers, etc.
---

<!--mdtogo:Short
    Initialize a new or existing kform working directory by creating initial files, initializing backend state, downloading modules, providers, etc.
-->

`init` initialize a new or existing kform working directory by creating initial files, initializing backend state, downloading modules, providers, etc.

### Synopsis

<!--mdtogo:Long-->

```
kform init [flags]
```

#### Args

```
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
# Initializes a new or existing kform working directory
$ kform init
```

<!--mdtogo-->
