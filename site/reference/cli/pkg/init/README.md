---
title: "`init`"
linkTitle: "init"
type: docs
description: >
  Initialize an empty provider/module package.
---

<!--mdtogo:Short
    Initialize an empty provider/module package.
-->

`init` initializes an existing empty directory as a module/provider package by adding a
kform.yaml, .kformignore and a placeholder `README.md` file.

### Synopsis

<!--mdtogo:Long-->

```
kform pkg init PACKAGE-TYPE DIR [flags]
```

#### Args

```
PACKAGE-TYPE:
  a package type can either be a provider or module
DIR:
  init fails if DIR does not already exist.
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
