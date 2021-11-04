# Podtato-head, Pulumi and Azure Container Apps

With the recent release of Azure Container Apps and Pulumi instant update of their Azure Native provider, I knew I had
to give it a spin.

Especially the Dapr part of Container Apps draw my attention. So here is a very short intro to all the parts of this
little demo.

## What is Pulumi?

Pulumi is an open source infrastructure as code tool for creating, deploying, and managing cloud infrastructure. Pulumi
works with traditional infrastructure like VMs, networks, and databases, in addition to modern architectures, including
containers, Kubernetes clusters, and serverless functions. Pulumi supports dozens of public, private, and hybrid cloud
service providers.

I use the golang, but there is support for JS/TS, Python, dotnet. I even saw a PR for Java, so plenty of choice for you
fellow dev. For more infos just hop onto their [website](https://www.pulumi.com/)

## Azure Native Provider

The Azure Native provider for Pulumi can be used to provision all the cloud resources available in Azure. It manages and
provisions resources using the Azure Resource Manager (ARM) APIs.

You can find the API docs to the [Azure Native Provider](https://www.pulumi.com/registry/packages/azure-native/)

## Podtato-head

We will deploy the 📨🚚 CNCF App Delivery SIG Demo [podtato-head](https://github.com/podtato-head/podtato-head)

Podtato-head demonstrates cloud-native application delivery scenarios using many different tools and services. It is
intended to help application delivery support teams test and decide which mechanism(s) to use.

## Azure Container Apps (KEDA, Dapr and envoy)

Microsoft released Azure Container Apps as a public preview during Ignite November 2021. Container Apps allows you to
run containerized applications on a serverless platform, No need to take care about the underlying infrastructure.

AKS acts as the control plane with and additional software stack such as:

- [Dapr](https://dapr.io/) helps developers build event-driven, resilient distributed applications. Whether on-premises,
  in the cloud, or on an edge device, Dapr helps you tackle the challenges that come with building microservices and
  keeps your code platform agnostic.

- [KEDA](https://keda.sh/) is a Kubernetes-based Event Driven Autoscaler. With KEDA, you can drive the scaling of any
  container in Kubernetes based on the number of events needing to be processed.

- [Envoy](https://www.envoyproxy.io/) is used to provide ingress functionality, traffic splitting for blue-green
  deployment and much more.

## Customize the Podtato-head main service

### Service invocation building block

Using service invocation, your application can reliably and securely communicate with other applications using the
standard gRPC or HTTP protocols.

Dapr uses a sidecar architecture. To invoke an application using Dapr, you use the invoke API on any Dapr instance. The
sidecar programming model encourages each applications to talk to its own instance of Dapr. The Dapr instances discover
and communicate with one another.

As we want to use the service invocation from dapr we need to add
the [go-client sdk ](https://docs.dapr.io/developing-applications/sdks/go/go-client/) to the Podtato-head main service.

```go
import (
dapr "github.com/dapr/go-sdk/client"
)
```

The invocation of the method happens with following snippet:

```go
resp, err := client.InvokeMethod(context.Background(), serviceMap[part], "images/"+image, "get")
if err != nil {
http.Error(w, err.Error(), http.StatusInternalServerError)
return
}
```

After the changes we build the new image of `podtato-head` and push them into the registry.

### Pulumi

The Pulumi part, is straight forward. Just create a Pulumi project
via `pulumi new  azure-go --dir pulumi-azure-container-apps` and follow the on-screen steps.

I just check, that the `go.mod` is up-to-date, and we are good to go.

The first part of the Pulumi program, is to init the Azure `Resource group`, the `Log Analytics workspace` and
the `Container App Environment`


Now we initialise the `Container App` for the parts of the `podtato-head` services. 

Important parts to mention are:

- We do not have a external ingress route `External:   pulumi.Bool(false),`
- We set the revision mode to multiple `ActiveRevisionsMode: pulumi.String("multiple"),`

Later will be quite handy to show the Split Traffic function from `Container Apps` via envoy

```go
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
                    "concurrentRequests": pulumi.String(20),
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
``` 

The main "body" part `podtato-head` services will now be declared.

Important part to mention is, We set external ingress route to true via  `External:   pulumi.Bool(true),` as we want to 
access the app from external.

```go
...
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
                "concurrentRequests": pulumi.String(20),
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
...
```

Now we can run `pulumi up` to see have the infrastructure gets deployed. Maybe you need to run first an `az login` to get
the credentials for your azure subscription.

```
Outputs:
    LatestRevisionFqdn                  : "https:/podtato-main--f6gz3h8.whitedesert-efebec22.northeurope.azurecontainerapps.io"
    LatestRevisionName                  : "podtato-main--f6gz3h8"
    podtato-hats_LatestRevisionName     : "podtato-hats--lhmqyw0"
    podtato-left-arm_LatestRevisionName : "podtato-left-arm--h5q7fns"
    podtato-left-leg_LatestRevisionName : "podtato-left-leg--dxfrjc1"
    podtato-right-arm_LatestRevisionName: "podtato-right-arm--tkxzyws"
    podtato-right-leg_LatestRevisionName: "podtato-right-leg--0dsjksm"
```

Keep an eye on this revision names. We will use them right now. If everything works well you should see our 
`podtato-head` man with a funny pirate head.

So let us change the tag of the head in our `main.go` to `v1` and re-apply the Pulumi program.

```go
...
{
    "podtato-hats",
    "v1",
},
...
```

The hat changed, and with the hat also the revision name. We can now split the traffic applied by assigning percentage 
values among the different revisions.

Add following properties to the hat struct

```go
web.TrafficWeightArgs{
  RevisionName: pulumi.String("podtato-hats--lhmqyw0"),
  Weight:       pulumi.IntPtr(50),
},
web.TrafficWeightArgs{
  LatestRevision: pulumi.BoolPtr(true),
  Weight:         pulumi.IntPtr(50),
},
```

And re-apply the Pulumi program. you should now see, when you refresh a that the traffic gets from the current service to
the old one. The effect are most of the time a new hat of `podtato-head` man

# Resources

- https://github.com/dirien/podtato-head
- https://github.com/dirien/infrastructure-as-code-workshop