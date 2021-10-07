# Terraform STACKIT


## Preinstall
```bash
brew tap hashicorp/tap
brew install hashicorp/tap/terraform
brew install golang
brew install cdktf
```

## Getting started
```bash
go build
cdktf get
cdktf synth

# ! create the credentials from **public** web project on STACKIT
# ! source **OpenStack RC** config file  

cdktf deploy
```


## More about the commands

Your cdktf go project is ready!

    cat help                Prints this message

Compile:

    go build              Builds your go project

Synthesize:

    cdktf synth [stack]   Synthesize Terraform resources to cdktf.out/

Diff:

    cdktf diff [stack]    Perform a diff (terraform plan) for the given stack

Deploy:

    cdktf deploy [stack]  Deploy the given stack

Destroy:

    cdktf destroy [stack] Destroy the given stack

Learn more about using modules and providers https://cdk.tf/modules-and-providers

