package main

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		release, err := helm.NewRelease(ctx, "pulumi-helm-operator", &helm.ReleaseArgs{
			Chart:           pulumi.String("oci://ghcr.io/pulumi/helm-charts/pulumi-kubernetes-operator"),
			Version:         pulumi.String("0.5.0"),
			CreateNamespace: pulumi.Bool(true),
			Name:            pulumi.String("pulumi-helm-operator"),
			Namespace:       pulumi.String("pulumi-kubernetes-operator"),
		})
		if err != nil {
			return err
		}

		err = DeployAzureMineCraftStack(ctx, release)
		if err != nil {
			return err
		}
		return nil
	})
}
