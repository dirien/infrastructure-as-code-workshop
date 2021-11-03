package main

import (
	"fmt"
	"github.com/pulumi/pulumi-azure-native/sdk/go/azure/operationalinsights"
	"github.com/pulumi/pulumi-azure-native/sdk/go/azure/resources"
	web "github.com/pulumi/pulumi-azure-native/sdk/go/azure/web/v20210301"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	resourceGroupName = "container-app-demo"
	workspaceName     = "workspace"
)

type podtatoPart struct {
	name    string
	version string
	traffic web.TrafficWeightArray
}

var podParts = []podtatoPart{
	{
		"podtato-hats",
		"v4",
		web.TrafficWeightArray{
			/*web.TrafficWeightArgs{
				RevisionName: pulumi.String("podtato-hats--d2kna50"),
				Weight:       pulumi.IntPtr(50),
			}*/
			web.TrafficWeightArgs{
				LatestRevision: pulumi.BoolPtr(true),
				Weight:         pulumi.IntPtr(100),
			},
		},
	},
	{
		"podtato-left-leg",
		"v1",
		web.TrafficWeightArray{
			web.TrafficWeightArgs{
				LatestRevision: pulumi.BoolPtr(true),
				Weight:         pulumi.IntPtr(100),
			},
		},
	},
	{
		"podtato-left-arm",
		"v2",
		web.TrafficWeightArray{
			web.TrafficWeightArgs{
				LatestRevision: pulumi.BoolPtr(true),
				Weight:         pulumi.IntPtr(100),
			},
		},
	},
	{
		"podtato-right-leg",
		"v1",
		web.TrafficWeightArray{
			web.TrafficWeightArgs{
				LatestRevision: pulumi.BoolPtr(true),
				Weight:         pulumi.IntPtr(100),
			},
		},
	},
	{
		"podtato-right-arm",
		"v3",
		web.TrafficWeightArray{
			web.TrafficWeightArgs{
				LatestRevision: pulumi.BoolPtr(true),
				Weight:         pulumi.IntPtr(100),
			},
		},
	},
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		resourceGroup, err := resources.NewResourceGroup(ctx, resourceGroupName, nil)

		workspace, err := operationalinsights.NewWorkspace(ctx, workspaceName, &operationalinsights.WorkspaceArgs{
			ResourceGroupName: resourceGroup.Name,
			RetentionInDays:   pulumi.Int(30),
			Sku: operationalinsights.WorkspaceSkuArgs{
				Name: pulumi.String("PerGB2018"),
			},
		})
		if err != nil {
			return err
		}

		sharedKeys := pulumi.All(resourceGroup.Name, workspace.Name).ApplyT(
			func(args []interface{}) (string, error) {
				resourceGroupName := args[0].(string)
				workspaceName := args[1].(string)
				accountKeys, err := operationalinsights.GetSharedKeys(ctx, &operationalinsights.GetSharedKeysArgs{
					ResourceGroupName: resourceGroupName,
					WorkspaceName:     workspaceName,
				})
				if err != nil {
					return "", err
				}

				return *accountKeys.PrimarySharedKey, nil
			},
		).(pulumi.StringOutput)

		kubeEnvironment, err := web.NewKubeEnvironment(ctx, "kubeEnvironment", &web.KubeEnvironmentArgs{
			ResourceGroupName:           resourceGroup.Name,
			Name:                        pulumi.String("kubeEnvironment"),
			Type:                        pulumi.String("managed"),
			InternalLoadBalancerEnabled: pulumi.Bool(false),
			AppLogsConfiguration: web.AppLogsConfigurationArgs{
				Destination: pulumi.String("log-analytics"),
				LogAnalyticsConfiguration: web.LogAnalyticsConfigurationArgs{
					CustomerId: workspace.CustomerId,
					SharedKey:  sharedKeys,
				},
			},
		})
		if err != nil {
			return err
		}

		for _, part := range podParts {
			podtatoHeadPart, err := web.NewContainerApp(ctx, part.name, &web.ContainerAppArgs{
				ResourceGroupName: resourceGroup.Name,
				Name:              pulumi.String(part.name),
				KubeEnvironmentId: kubeEnvironment.ID(),
				Configuration: web.ConfigurationArgs{
					ActiveRevisionsMode: pulumi.String("multiple"),
					Ingress: web.IngressArgs{
						External:   pulumi.Bool(false),
						TargetPort: pulumi.IntPtr(9000),
						Traffic:    part.traffic,
					},
				},
				Template: web.TemplateArgs{
					Containers: web.ContainerArray{
						web.ContainerArgs{
							Name:  pulumi.String(part.name),
							Image: pulumi.String(fmt.Sprintf("dirien/%s:%s", part.name, part.version)),
						},
					},
					Scale: web.ScaleArgs{
						MaxReplicas: pulumi.IntPtr(5),
						MinReplicas: pulumi.IntPtr(0),
						Rules: web.ScaleRuleArray{
							web.ScaleRuleArgs{
								Name: pulumi.String("http-rule"),
								Http: web.HttpScaleRuleArgs{
									Metadata: pulumi.StringMap{
										"concurrentRequests": pulumi.String("20"),
									},
								},
							},
						},
					},
					Dapr: web.DaprArgs{
						Enabled: pulumi.BoolPtr(true),
						AppId:   pulumi.String(part.name),
						AppPort: pulumi.IntPtr(9000),
					},
				},
			})
			if err != nil {
				return err
			}
			ctx.Export(fmt.Sprintf("%s_LatestRevisionName", part.name), podtatoHeadPart.LatestRevisionName)
		}

		mainApp, err := web.NewContainerApp(ctx, "podtato-main", &web.ContainerAppArgs{
			ResourceGroupName: resourceGroup.Name,
			Name:              pulumi.String("podtato-main"),
			KubeEnvironmentId: kubeEnvironment.ID(),
			Configuration: web.ConfigurationArgs{
				Ingress: web.IngressArgs{
					External:   pulumi.Bool(true),
					TargetPort: pulumi.IntPtr(9000),
				},
			},
			Template: web.TemplateArgs{
				Containers: web.ContainerArray{
					web.ContainerArgs{
						Name:  pulumi.String("podtato-main"),
						Image: pulumi.String("dirien/podtato-main:v1@sha256:671a7776e448cdf5de5b01a44812d29137a9e6f9267be1fc919b2d59f69040e7"),
					},
				},
				Scale: web.ScaleArgs{
					MaxReplicas: pulumi.IntPtr(5),
					MinReplicas: pulumi.IntPtr(0),
					Rules: web.ScaleRuleArray{
						web.ScaleRuleArgs{
							Name: pulumi.String("http-rule"),
							Http: web.HttpScaleRuleArgs{
								Metadata: pulumi.StringMap{
									"concurrentRequests": pulumi.String("20"),
								},
							},
						},
					},
				},
				Dapr: web.DaprArgs{
					Enabled: pulumi.BoolPtr(true),
					AppId:   pulumi.String("podtato-main"),
					AppPort: pulumi.IntPtr(9000),
				},
			},
		})

		if err != nil {
			return err
		}

		ctx.Export("LatestRevisionFqdn", pulumi.Sprintf("https:/%s", mainApp.LatestRevisionFqdn))
		ctx.Export("LatestRevisionName", mainApp.LatestRevisionName)

		return nil
	})
}
