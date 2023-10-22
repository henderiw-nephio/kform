---
title: "`tree`"
linkTitle: "tree"
type: docs
description: >
  Display resources, files and packages in a tree structure.
---

<!--mdtogo:Short
    Display resources, files and packages in a tree structure.
-->

`tree` initializes an existing empty directory as a module/provider package by adding a
kform.yaml, .kformignore and a placeholder `README.md` file.

### Synopsis

<!--mdtogo:Long-->

```
kform pkg tree [DIR] [flags]
```

#### Args

```
DIR:
  tree uses the current directory if no DIR is supplied
```

#### Flags

```
--description
  Short description of the package. (default "sample description")
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
# Creates a new module package in the module-hello-world directory.
$ mkdir module-hello-world; kform pkg init module module-hello-world \\
    --description "my hello-world module implementation"
```

```shell
# Creates a new provider package in the provider-hello-world directory.
$ mkdir provider-hello-world; kform pkg init provider provider-hello-world \\
    --description "my hello-world provider implementation"
```

<!--mdtogo-->
