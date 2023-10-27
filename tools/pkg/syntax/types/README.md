# terraform block types

# terraform

- backend
    used to store state -> default is local
    mutually exclusive with cloud
    examples: remote, s3, azurerm, consul
    config is specific to the type
- cloud
    config:
        organization
        hostname
        workspaces
- required_providers
    name
    source
    aliases
- provider-meta
    seen as experiemental -> example is to use it to gather statistics on the provider usage

# provider
    - config is based on the provider schema
    - expressions: use variables but no resources/data can be used
    - meta-arguments
        alias
        version -> depreciated
        source/count/for_each/depends_on -> not expected

# variable
    - like function arguments
    - meta-arguments
        default
        type (string, number, bool, list<TYPE>, set<TYPE>, map<TYPE>, object)
        description
        validation
        sensitive -> 
        nullable
    - variables can be assigned using -var or .tfvars files
    - idea is to use KRM here
    - cannot have expressions

# locals
    - like local variables -> can use expressions
    - can have expressions

# output
    - like function outputs -> expose variables to the other module
    - meta-arguments
        - value -> expression
        - description
        - sensitive
        - depends_on
        - precondition
        - postcondition

# module
    - meta-arguments:
       - version
       - source = "./aws_vpc" (mandatory)
            -> europe-docker.pkg.dev/srlinux/eu.gcr.io/<TYPE>:<VERSION>
            -> [<HOSTNAME>/]<NAMESPACE>/<TYPE>
                HOSTNAME: registry.terraform.io
                NAMESPACE: An organizational namespace within the specified registry
                TYPE: typically the provider 
            -> ./,,, -> local path
            -> hashicorp/conusl/aws -> registry (app.terraform.io/example-corp/k8s-cluster/azurerm)
            supported:
                - local
                - registry
                - github/bitbucket/git
                - s3
                - gcs
       - providers = { aws = aws.west }
       - for_each
       - count
       - depends_on
       - <dynamic input variables>

# resource
    - meta-arguments:
        provider
        count
        for_each mutually exclusive with count
        depends_on
        lifecycle
            create_before_destroy
            prevent_destroy
            replace_triggered_by
            ignore_changes
        precondition
        postcondition
        connection
        provisioner

# data
    - meta-arguments:
        -> see resource

# moved
    - to deal with module resource changes
    - meta-arguments:
        from
        to

# import
    - Use the import block to import existing infrastructure resources into Terraform, bringing them under Terraform's management
    - Not sure if we need this as a resource will do this for us

# check


