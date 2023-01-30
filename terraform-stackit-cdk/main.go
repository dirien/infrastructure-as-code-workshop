package main

import (
	"fmt"
	"log"
	"os"

	"cdk.tf/go/stack/generated/hashicorp/cloudinit/datacloudinitconfig"
	initProvider "cdk.tf/go/stack/generated/hashicorp/cloudinit/provider"
	"cdk.tf/go/stack/generated/terraform-provider-openstack/openstack/computefloatingipassociatev2"
	"cdk.tf/go/stack/generated/terraform-provider-openstack/openstack/computeinstancev2"
	"cdk.tf/go/stack/generated/terraform-provider-openstack/openstack/computekeypairv2"
	"cdk.tf/go/stack/generated/terraform-provider-openstack/openstack/dataopenstackimagesimagev2"
	"cdk.tf/go/stack/generated/terraform-provider-openstack/openstack/networkingfloatingipv2"
	"cdk.tf/go/stack/generated/terraform-provider-openstack/openstack/networkingnetworkv2"
	"cdk.tf/go/stack/generated/terraform-provider-openstack/openstack/networkingrouterinterfacev2"
	"cdk.tf/go/stack/generated/terraform-provider-openstack/openstack/networkingrouterv2"
	"cdk.tf/go/stack/generated/terraform-provider-openstack/openstack/networkingsecgrouprulev2"
	"cdk.tf/go/stack/generated/terraform-provider-openstack/openstack/networkingsecgroupv2"
	"cdk.tf/go/stack/generated/terraform-provider-openstack/openstack/networkingsubnetv2"
	openstackProvider "cdk.tf/go/stack/generated/terraform-provider-openstack/openstack/provider"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

func MinecraftStack(scope constructs.Construct, id string) cdktf.TerraformStack {
	stack := cdktf.NewTerraformStack(scope, &id)

	pubKeyFile, err := os.ReadFile("../ssh/workshop.pub")
	if err != nil {
		log.Fatal(err)
	}

	initProvider.NewCloudinitProvider(stack, jsii.String("cloudinit-provider"), &initProvider.CloudinitProviderConfig{})

	openstackProvider.NewOpenstackProvider(stack, jsii.String("openstack-provider"), &openstackProvider.OpenstackProviderConfig{})

	keypair := computekeypairv2.NewComputeKeypairV2(stack, jsii.String("keypair"), &computekeypairv2.ComputeKeypairV2Config{
		Name:      jsii.String(fmt.Sprintf("%s-kp", id)),
		PublicKey: jsii.String(string(pubKeyFile)),
	})

	network := networkingnetworkv2.NewNetworkingNetworkV2(stack, jsii.String("network"), &networkingnetworkv2.NetworkingNetworkV2Config{
		Name:         jsii.String(fmt.Sprintf("%s-net", id)),
		AdminStateUp: jsii.Bool(true),
	})

	subnet := networkingsubnetv2.NewNetworkingSubnetV2(stack, jsii.String("subnet"), &networkingsubnetv2.NetworkingSubnetV2Config{
		Name:           jsii.String(fmt.Sprintf("%s-snet", id)),
		NetworkId:      network.Id(),
		Cidr:           jsii.String("10.1.10.0/24"),
		IpVersion:      jsii.Number(4),
		DnsNameservers: jsii.Strings("8.8.8.8", "8.8.4.4"),
	})

	floating := networkingnetworkv2.NewNetworkingNetworkV2(stack, jsii.String("floating"), &networkingnetworkv2.NetworkingNetworkV2Config{
		Name: jsii.String("floating-net"),
	})

	router := networkingrouterv2.NewNetworkingRouterV2(stack, jsii.String("router"), &networkingrouterv2.NetworkingRouterV2Config{
		Name:              jsii.String(fmt.Sprintf("%s-router", id)),
		AdminStateUp:      jsii.Bool(true),
		ExternalNetworkId: floating.Id(),
	})

	networkingrouterinterfacev2.NewNetworkingRouterInterfaceV2(stack, jsii.String("ri"), &networkingrouterinterfacev2.NetworkingRouterInterfaceV2Config{
		RouterId: router.Id(),
		SubnetId: subnet.Id(),
	})

	secgoup := networkingsecgroupv2.NewNetworkingSecgroupV2(stack, jsii.String("sg"), &networkingsecgroupv2.NetworkingSecgroupV2Config{
		Name:        jsii.String(fmt.Sprintf("%s-sec", id)),
		Description: jsii.String("Security group for the Terraform minecraft instances"),
	})

	networkingsecgrouprulev2.NewNetworkingSecgroupRuleV2(stack, jsii.String("sgr-22"), &networkingsecgrouprulev2.NetworkingSecgroupRuleV2Config{
		Direction:       jsii.String("ingress"),
		Ethertype:       jsii.String("IPv4"),
		Protocol:        jsii.String("tcp"),
		PortRangeMin:    jsii.Number(22),
		PortRangeMax:    jsii.Number(22),
		RemoteIpPrefix:  jsii.String("0.0.0.0/0"),
		SecurityGroupId: secgoup.Id(),
	})

	networkingsecgrouprulev2.NewNetworkingSecgroupRuleV2(stack, jsii.String("sgr-25565"), &networkingsecgrouprulev2.NetworkingSecgroupRuleV2Config{
		Direction:       jsii.String("ingress"),
		Ethertype:       jsii.String("IPv4"),
		Protocol:        jsii.String("tcp"),
		PortRangeMin:    jsii.Number(25565),
		PortRangeMax:    jsii.Number(25565),
		RemoteIpPrefix:  jsii.String("0.0.0.0/0"),
		SecurityGroupId: secgoup.Id(),
	})

	cloudInit, err := os.ReadFile("../config/cloud-init.yaml")
	if err != nil {
		log.Fatal(err)
	}
	configpart := []*datacloudinitconfig.DataCloudinitConfigPart{{
		ContentType: jsii.String("text/cloud-config"),
		Content:     jsii.String(string(cloudInit)),
	}}

	cloudinit := datacloudinitconfig.NewDataCloudinitConfig(stack, jsii.String("ubuntu-config"), &datacloudinitconfig.DataCloudinitConfigConfig{
		Gzip:         jsii.Bool(true),
		Base64Encode: jsii.Bool(true),
		Part:         &configpart,
	})

	activeImage := dataopenstackimagesimagev2.NewDataOpenstackImagesImageV2(stack, jsii.String("image"), &dataopenstackimagesimagev2.DataOpenstackImagesImageV2Config{
		Name: jsii.String("Ubuntu 20.04"),
		Properties: &map[string]*string{
			"Status": jsii.String("active"),
		},
		//MostRecent: jsii.Bool(true),
	})

	vm := computeinstancev2.NewComputeInstanceV2(stack, jsii.String("vm"), &computeinstancev2.ComputeInstanceV2Config{
		Name:           jsii.String(fmt.Sprintf("%s-ubuntu", id)),
		FlavorName:     jsii.String("c1.3"),
		KeyPair:        keypair.Name(),
		SecurityGroups: jsii.Strings("default", *secgoup.Name()),
		UserData:       cloudinit.Rendered(),
		Network: &[]*computeinstancev2.ComputeInstanceV2Network{
			{Name: network.Name()},
		},
		BlockDevice: &[]*computeinstancev2.ComputeInstanceV2BlockDevice{
			{
				Uuid:                activeImage.Id(),
				SourceType:          jsii.String("image"),
				BootIndex:           jsii.Number(0),
				DestinationType:     jsii.String("volume"),
				VolumeSize:          jsii.Number(10),
				DeleteOnTermination: jsii.Bool(true),
			},
		},
	})
	fip := networkingfloatingipv2.NewNetworkingFloatingipV2(stack, jsii.String("minecraft-fip"), &networkingfloatingipv2.NetworkingFloatingipV2Config{
		Pool: jsii.String("floating-net"),
	})
	computefloatingipassociatev2.NewComputeFloatingipAssociateV2(stack, jsii.String("minecraft-fipa"), &computefloatingipassociatev2.ComputeFloatingipAssociateV2Config{
		InstanceId: vm.Id(),
		FloatingIp: fip.Address(),
	})

	cdktf.NewTerraformOutput(stack, jsii.String("minecraft-public"), &cdktf.TerraformOutputConfig{
		Value:       fip.Address(),
		Description: jsii.String("The public ips of the nodes"),
	})

	return stack
}

func main() {
	app := cdktf.NewApp(nil)

	MinecraftStack(app, "minecraft-cdk")

	app.Synth()
}
