package main

import (
	"github.com/pulumi/pulumi-azure-native/sdk/go/azure/containerregistry"
	"github.com/pulumi/pulumi-azure-native/sdk/go/azure/operationalinsights"
	"github.com/pulumi/pulumi-azure-native/sdk/go/azure/resources"
	"github.com/pulumi/pulumi-azure-native/sdk/go/azure/web"
	"github.com/pulumi/pulumi-docker/sdk/v3/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	resourceGroupName = "container-app-demo"
	workspaceName     = "workspace"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		resourceGroup, err := resources.NewResourceGroup(ctx, resourceGroupName, nil)

		workspace, err := operationalinsights.NewWorkspace(ctx, workspaceName, &operationalinsights.WorkspaceArgs{
			ResourceGroupName: resourceGroup.Name,
			RetentionInDays:   pulumi.Int(30),
			Sku: operationalinsights.WorkspaceSkuArgs{
				Name: pulumi.String("PerGB2018"),
			},
		}, pulumi.DependsOn([]pulumi.Resource{resourceGroup}))
		if err != nil {
			return err
		}

		sharedKeys, err := operationalinsights.GetSharedKeys(ctx, &operationalinsights.GetSharedKeysArgs{
			ResourceGroupName: resourceGroupName,
			WorkspaceName:     workspaceName,
		})
		if err != nil {
			return err
		}

		kubeEnvironment, err := web.NewKubeEnvironment(ctx, "kubeEnvironment", &web.KubeEnvironmentArgs{
			ResourceGroupName:           resourceGroup.Name,
			Name:                        pulumi.String("kubeEnvironment"),
			InternalLoadBalancerEnabled: pulumi.Bool(false),
			AppLogsConfiguration: web.AppLogsConfigurationArgs{
				Destination: pulumi.String("log-analytics"),
				LogAnalyticsConfiguration: web.LogAnalyticsConfigurationArgs{
					CustomerId: workspace.CustomerId,
					SharedKey:  pulumi.StringPtr(*sharedKeys.PrimarySharedKey),
				},
			},
		})
		if err != nil {
			return err
		}

		registry, err := containerregistry.NewRegistry(ctx, "registry", &containerregistry.RegistryArgs{
			ResourceGroupName: resourceGroup.Name,
			Sku: containerregistry.SkuArgs{
				Name: pulumi.String("Basic"),
			},
			AdminUserEnabled: pulumi.Bool(true),
		})
		if err != nil {
			return err
		}
		registryCredentials, err := containerregistry.ListRegistryCredentials(ctx, &containerregistry.ListRegistryCredentialsArgs{
			ResourceGroupName: resourceGroupName,
			RegistryName:      "registry",
		})
		if err != nil {
			return err
		}
		newImage, err := docker.NewImage(ctx, "image", &docker.ImageArgs{
			ImageName: pulumi.Sprintf("%s/app:v1.0.0", registry.LoginServer),
			Build: docker.DockerBuildArgs{
				Context: pulumi.String("/app"),
			},
			Registry: docker.ImageRegistryArgs{
				Server:   registry.LoginServer,
				Username: pulumi.Sprintf("%s", registryCredentials.Username),
				Password: pulumi.Sprintf("%s", registryCredentials.Passwords[0].Name),
			},
		})
		if err != nil {
			return err
		}

		containerApp, err := web.NewContainerApp(ctx, "app", &web.ContainerAppArgs{
			ResourceGroupName: resourceGroup.Name,
			KubeEnvironmentId: kubeEnvironment.ID(),
			Configuration: web.ConfigurationArgs{
				Ingress: web.IngressArgs{
					External:   pulumi.Bool(true),
					TargetPort: pulumi.IntPtr(80),
				},
				Registries: web.RegistryCredentialsArray{
					web.RegistryCredentialsArgs{
						Server:            registry.LoginServer,
						Username:          pulumi.Sprintf("%s", registryCredentials.Username),
						PasswordSecretRef: pulumi.String("pwd")},
				},
				Secrets: web.SecretArray{
					web.SecretArgs{
						Name:  pulumi.String("pwd"),
						Value: pulumi.Sprintf("%s", registryCredentials.Passwords[0].Name),
					},
				},
			},
			Template: web.TemplateArgs{
				Containers: web.ContainerArray{
					web.ContainerArgs{
						Name:  pulumi.String("myapp"),
						Image: newImage.ImageName,
					},
				},
			},
		})
		if err != nil {
			return err
		}

		ctx.Export("url", pulumi.Sprintf("https:/%s", containerApp.LatestRevisionFqdn))

		return nil
	})
}
