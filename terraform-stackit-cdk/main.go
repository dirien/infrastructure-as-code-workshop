package main

import (
	"cdk.tf/go/stack/generated/hashicorp/cloudinit"
	"cdk.tf/go/stack/generated/terraform-provider-openstack/openstack"
	"fmt"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"io/ioutil"
	"log"
)

func MinecraftStack(scope constructs.Construct, id string) cdktf.TerraformStack {
	stack := cdktf.NewTerraformStack(scope, &id)

	pubKeyFile, err := ioutil.ReadFile("../ssh/workshop.pub")
	if err != nil {
		log.Fatal(err)
	}

	cloudinit.NewCloudinitProvider(stack, jsii.String("cloudinit-provider"), &cloudinit.CloudinitProviderConfig{})

	openstack.NewOpenstackProvider(stack, jsii.String("openstack-provider"), &openstack.OpenstackProviderConfig{})

	keypair := openstack.NewComputeKeypairV2(stack, jsii.String("keypair"), &openstack.ComputeKeypairV2Config{
		Name:      jsii.String(fmt.Sprintf("%s-kp", id)),
		PublicKey: jsii.String(string(pubKeyFile)),
	})

	network := openstack.NewNetworkingNetworkV2(stack, jsii.String("network"), &openstack.NetworkingNetworkV2Config{
		Name:         jsii.String(fmt.Sprintf("%s-net", id)),
		AdminStateUp: jsii.Bool(true),
	})

	subnet := openstack.NewNetworkingSubnetV2(stack, jsii.String("subnet"), &openstack.NetworkingSubnetV2Config{
		Name:           jsii.String(fmt.Sprintf("%s-snet", id)),
		NetworkId:      network.Id(),
		Cidr:           jsii.String("10.1.10.0/24"),
		IpVersion:      jsii.Number(4),
		DnsNameservers: jsii.Strings("8.8.8.8", "8.8.4.4"),
	})

	floating := openstack.NewDataOpenstackNetworkingNetworkV2(stack, jsii.String("floating"), &openstack.DataOpenstackNetworkingNetworkV2Config{
		Name: jsii.String("floating-net"),
	})

	router := openstack.NewNetworkingRouterV2(stack, jsii.String("router"), &openstack.NetworkingRouterV2Config{
		Name:              jsii.String(fmt.Sprintf("%s-router", id)),
		AdminStateUp:      jsii.Bool(true),
		ExternalNetworkId: floating.Id(),
	})

	openstack.NewNetworkingRouterInterfaceV2(stack, jsii.String("ri"), &openstack.NetworkingRouterInterfaceV2Config{
		RouterId: router.Id(),
		SubnetId: subnet.Id(),
	})

	secgoup := openstack.NewNetworkingSecgroupV2(stack, jsii.String("sg"), &openstack.NetworkingSecgroupV2Config{
		Name:        jsii.String(fmt.Sprintf("%s-sec", id)),
		Description: jsii.String("Security group for the Terraform minecraft instances"),
	})

	openstack.NewNetworkingSecgroupRuleV2(stack, jsii.String("sgr-22"), &openstack.NetworkingSecgroupRuleV2Config{
		Direction:       jsii.String("ingress"),
		Ethertype:       jsii.String("IPv4"),
		Protocol:        jsii.String("tcp"),
		PortRangeMin:    jsii.Number(22),
		PortRangeMax:    jsii.Number(22),
		RemoteIpPrefix:  jsii.String("0.0.0.0/0"),
		SecurityGroupId: secgoup.Id(),
	})

	openstack.NewNetworkingSecgroupRuleV2(stack, jsii.String("sgr-25565"), &openstack.NetworkingSecgroupRuleV2Config{
		Direction:       jsii.String("ingress"),
		Ethertype:       jsii.String("IPv4"),
		Protocol:        jsii.String("tcp"),
		PortRangeMin:    jsii.Number(25565),
		PortRangeMax:    jsii.Number(25565),
		RemoteIpPrefix:  jsii.String("0.0.0.0/0"),
		SecurityGroupId: secgoup.Id(),
	})

	cloudInit, err := ioutil.ReadFile("../config/cloud-init.yaml")
	if err != nil {
		log.Fatal(err)
	}
	configpart := []*cloudinit.DataCloudinitConfigPart{{
		ContentType: jsii.String("text/cloud-config"),
		Content:     jsii.String(string(cloudInit)),
	}}

	cloudinit := cloudinit.NewDataCloudinitConfig(stack, jsii.String("ubuntu-config"), &cloudinit.DataCloudinitConfigConfig{
		Gzip:         jsii.Bool(true),
		Base64Encode: jsii.Bool(true),
		Part:         &configpart,
	})

	activeImage := openstack.NewDataOpenstackImagesImageV2(stack, jsii.String("image"), &openstack.DataOpenstackImagesImageV2Config{
		Name: jsii.String("Ubuntu 20.04"),
		Properties: &map[string]*string{
			"Status": jsii.String("active"),
		},
		//MostRecent: jsii.Bool(true),
	})

	vm := openstack.NewComputeInstanceV2(stack, jsii.String("vm"), &openstack.ComputeInstanceV2Config{
		Name:           jsii.String(fmt.Sprintf("%s-ubuntu", id)),
		FlavorName:     jsii.String("c1.3"),
		KeyPair:        keypair.Name(),
		SecurityGroups: jsii.Strings("default", *secgoup.Name()),
		UserData:       cloudinit.Rendered(),
		Network: &[]*openstack.ComputeInstanceV2Network{
			{Name: network.Name()},
		},
		BlockDevice: &[]*openstack.ComputeInstanceV2BlockDevice{
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

	fip := openstack.NewNetworkingFloatingipV2(stack, jsii.String("minecraft-fip"), &openstack.NetworkingFloatingipV2Config{
		Pool: jsii.String("floating-net"),
	})

	openstack.NewComputeFloatingipAssociateV2(stack, jsii.String("minecraft-fipa"), &openstack.ComputeFloatingipAssociateV2Config{
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
