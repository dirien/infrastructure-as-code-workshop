package main

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func DeployAzureMineCraftStack(ctx *pulumi.Context, resource pulumi.Resource) error {
	config := config.New(ctx, "")
	pulumiAccessToken := config.Require("pulumiAccessToken")
	armClientID := config.Require("armClientID")
	armClientSecret := config.Require("armClientSecret")
	armTenantID := config.Require("armTenantID")
	armSubscriptionID := config.Require("armSubscriptionID")

	accessToken, err := corev1.NewSecret(ctx, "accesstoken", &corev1.SecretArgs{
		StringData: pulumi.StringMap{"accessToken": pulumi.String(pulumiAccessToken)},
	})
	if err != nil {
		return err
	}
	azureCreds, err := corev1.NewSecret(ctx, "azure-creds", &corev1.SecretArgs{
		Metadata: metav1.ObjectMetaPtr(&metav1.ObjectMetaArgs{
			Name: pulumi.String("azure-creds"),
		}).ToObjectMetaPtrOutput(),
		StringData: pulumi.StringMap{
			"ARM_CLIENT_ID":       pulumi.String(armClientID),
			"ARM_CLIENT_SECRET":   pulumi.String(armClientSecret),
			"ARM_TENANT_ID":       pulumi.String(armTenantID),
			"ARM_SUBSCRIPTION_ID": pulumi.String(armSubscriptionID),
		},
	})
	if err != nil {
		return err
	}

	// Create an S3 bucket through the operator
	_, err = apiextensions.NewCustomResource(ctx, "azure-minecraft-stack",
		&apiextensions.CustomResourceArgs{
			Metadata: metav1.ObjectMetaPtr(&metav1.ObjectMetaArgs{
				Name: pulumi.String("azure-minecraft"),
			}).ToObjectMetaPtrOutput(),
			ApiVersion: pulumi.String("pulumi.com/v1"),
			Kind:       pulumi.String("Stack"),
			OtherFields: kubernetes.UntypedArgs{
				"spec": map[string]interface{}{
					"stack":             "dirien/pulumi-azure-go/dev",
					"projectRepo":       "https://github.com/dirien/infrastructure-as-code-workshop",
					"branch":            "refs/remotes/origin/main",
					"repoDir":           "pulumi-azure-go",
					"accessTokenSecret": accessToken.Metadata.Name(),
					"config": map[string]string{
						"azure-native:location": "westeurope",
						"pulumi-azure-go:name":  "minecraft",
					},
					"envSecrets":        []interface{}{azureCreds.Metadata.Name()},
					"destroyOnFinalize": true,
				},
			},
		}, pulumi.DependsOn([]pulumi.Resource{accessToken, azureCreds, resource}))
	return nil
}
