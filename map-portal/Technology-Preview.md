## Technology Preview

This feature is currently in the technical preview stage, and related features may change in future versions and are not enabled by default. 

For trial, you can enable as below:
* change featureGate in src/environments/environment.prod.ts
```
{
  ......
  features: {
    appMgmt: true
  }
}
```
* build map-console image    
make build

* Restart map-console with the new image
