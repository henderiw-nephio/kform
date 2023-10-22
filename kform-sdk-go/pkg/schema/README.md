## provider config load

When the terraform runtime executes the configuration file, it loads the configuration provided in the provider and 


c := terraform.NewResourceConfigRaw(tc.Config) -> retruns terraform.ResourceConfig
diags := tc.P.Configure(context.Background(), c)