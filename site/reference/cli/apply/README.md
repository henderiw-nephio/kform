---
title: "`apply`"
linkTitle: "apply"
type: docs
description: >
  Creates or updates KRM resources according to kform configuration files in the current directory.

  By default, kform will generate a new plan and present it for your approval before taking any action. You can optionally apply the KRM resources with auto-approval
---

<!--mdtogo:Short
    Creates or updates KRM resources according to kform configuration files in the current directory.

  By default, kform will generate a new plan and present it for your approval before taking any action. You can optionally apply the KRM resources with auto-approval
-->

`apply` creates or updates KRM resources according to kform configuration files in the current directory.

By default, kform will generate a new plan and present it for your approval before taking any action. You can optionally apply the KRM resources with auto-approval

### Synopsis

<!--mdtogo:Long-->

```
kform apply [flags]
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
# Creates or updates KRM resources according to kform configuration files in the current directory
$ kform apply
```

<!--mdtogo-->
