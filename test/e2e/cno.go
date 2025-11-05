package e2e

import (
	"context"
	"fmt"
	"path/filepath"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"

	//exutil "github.com/openshift/openshift-network-operator/test/util"
	exutil "github.com/openshift/origin/test/extended/util"
	compat_otp "github.com/openshift/origin/test/extended/util/compat_otp"
	e2enode "k8s.io/kubernetes/test/e2e/framework/node"
)

var _ = g.Describe("[sig-network] CNO", func() {

	defer g.GinkgoRecover()
	var oc *compat_otp.CLI

	g.It("Author:anusaxen-High-73205-High-72817-Make sure internalJoinSubnet and internalTransitSwitchSubnet is configurable post install as a Day 2 operation [Disruptive]", func() {
		var (
			buildPruningBaseDir    = compat_otp.FixturePath("testdata", "networking")
			pingPodNodeTemplate    = filepath.Join(buildPruningBaseDir, "ping-for-pod-specific-node-template.yaml")
			genericServiceTemplate = filepath.Join(buildPruningBaseDir, "service-generic-template.yaml")
		)
		ipStackType := checkIPStackType(oc)
		o.Expect(ipStackType).NotTo(o.BeEmpty())

		nodeList, err := e2enode.GetReadySchedulableNodes(context.TODO(), oc.KubeFramework().ClientSet)
		o.Expect(err).NotTo(o.HaveOccurred())
		if len(nodeList.Items) < 2 {
			g.Skip("This case requires 2 nodes, but the cluster has less than two nodes")
		}
		compat_otp.By("create a hello pod1 in namespace")
		pod1ns := pingPodResourceNode{
			name:      "hello-pod1",
			namespace: oc.Namespace(),
			nodename:  nodeList.Items[0].Name,
			template:  pingPodNodeTemplate,
		}
		pod1ns.createPingPodNode(oc)
		waitPodReady(oc, oc.Namespace(), pod1ns.name)

		exutil.By("create a hello-pod2 in namespace")
		pod2ns := pingPodResourceNode{
			name:      "hello-pod2",
			namespace: oc.Namespace(),
			nodename:  nodeList.Items[1].Name,
			template:  pingPodNodeTemplate,
		}
		pod2ns.createPingPodNode(oc)
		waitPodReady(oc, oc.Namespace(), pod2ns.name)

		g.By("Create a test service backing up both the above pods")
		svc := genericServiceResource{
			servicename:           "test-service-73205",
			namespace:             oc.Namespace(),
			protocol:              "TCP",
			selector:              "hello-pod",
			serviceType:           "ClusterIP",
			ipFamilyPolicy:        "",
			internalTrafficPolicy: "Cluster",
			externalTrafficPolicy: "",
			template:              genericServiceTemplate,
		}
		if ipStackType == "ipv4single" {
			svc.ipFamilyPolicy = "SingleStack"
		} else {
			svc.ipFamilyPolicy = "PreferDualStack"
		}
		svc.createServiceFromParams(oc)
		//custom patches to test depending on type of cluster addressing
		customPatchIPv4 := "{\"spec\":{\"defaultNetwork\":{\"ovnKubernetesConfig\":{\"ipv4\":{\"internalJoinSubnet\": \"100.99.0.0/16\",\"internalTransitSwitchSubnet\": \"100.69.0.0/16\"}}}}}"
		customPatchIPv6 := "{\"spec\":{\"defaultNetwork\":{\"ovnKubernetesConfig\":{\"ipv6\":{\"internalJoinSubnet\": \"ab98::/64\",\"internalTransitSwitchSubnet\": \"ab97::/64\"}}}}}"
		customPatchDualstack := "{\"spec\":{\"defaultNetwork\":{\"ovnKubernetesConfig\":{\"ipv4\":{\"internalJoinSubnet\": \"100.99.0.0/16\",\"internalTransitSwitchSubnet\": \"100.69.0.0/16\"},\"ipv6\": {\"internalJoinSubnet\": \"ab98::/64\",\"internalTransitSwitchSubnet\": \"ab97::/64\"}}}}}"

		//gather original cluster values so that we can defer to them later once test done
		currentinternalJoinSubnetIPv4Value, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("Network.operator.openshift.io/cluster", "-o=jsonpath={.items[*].spec.defaultNetwork.ovnKubernetesConfig.ipv4.internalJoinSubnet}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		currentinternalTransitSwSubnetIPv4Value, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("Network.operator.openshift.io/cluster", "-o=jsonpath={.items[*].spec.defaultNetwork.ovnKubernetesConfig.ipv4.internalTransitSwitchSubnet}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		currentinternalJoinSubnetIPv6Value, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("Network.operator.openshift.io/cluster", "-o=jsonpath={.items[*].spec.defaultNetwork.ovnKubernetesConfig.ipv6.internalJoinSubnet}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
		currentinternalTransitSwSubnetIPv6Value, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("Network.operator.openshift.io/cluster", "-o=jsonpath={.items[*].spec.defaultNetwork.ovnKubernetesConfig.ipv6.internalTransitSwitchSubnet}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())

		//if any of value is null on exisiting cluster, it indicates that cluster came up with following default values assigned by OVNK
		if (currentinternalJoinSubnetIPv4Value == "") || (currentinternalJoinSubnetIPv6Value == "") {
			currentinternalJoinSubnetIPv4Value = "100.64.0.0/16"
			currentinternalTransitSwSubnetIPv4Value = "100.88.0.0/16"
			currentinternalJoinSubnetIPv6Value = "fd98::/64"
			currentinternalTransitSwSubnetIPv6Value = "fd97::/64"
		}

		//vars to patch cluster back to original state
		patchIPv4original := "{\"spec\":{\"defaultNetwork\":{\"ovnKubernetesConfig\":{\"ipv4\":{\"internalJoinSubnet\": \"" + currentinternalJoinSubnetIPv4Value + "\",\"internalTransitSwitchSubnet\": \"" + currentinternalTransitSwSubnetIPv4Value + "\"}}}}}"
		patchIPv6original := "{\"spec\":{\"defaultNetwork\":{\"ovnKubernetesConfig\":{\"ipv6\":{\"internalJoinSubnet\": \"" + currentinternalJoinSubnetIPv6Value + "\",\"internalTransitSwitchSubnet\": \"" + currentinternalTransitSwSubnetIPv6Value + "\"}}}}}"
		patchDualstackoriginal := "{\"spec\":{\"defaultNetwork\":{\"ovnKubernetesConfig\":{\"ipv4\":{\"internalJoinSubnet\": \"" + currentinternalJoinSubnetIPv4Value + "\",\"internalTransitSwitchSubnet\": \"" + currentinternalTransitSwSubnetIPv4Value + "\"},\"ipv6\": {\"internalJoinSubnet\": \"" + currentinternalJoinSubnetIPv6Value + "\",\"internalTransitSwitchSubnet\": \"" + currentinternalTransitSwSubnetIPv6Value + "\"}}}}}"

		if ipStackType == "ipv4single" {
			defer func() {
				patchResourceAsAdmin(oc, "Network.operator.openshift.io/cluster", patchIPv4original)
				err := checkOVNKState(oc)
				exutil.AssertWaitPollNoErr(err, fmt.Sprintf("OVNkube didn't trigger or rolled out successfully post oc patch"))
			}()
			patchResourceAsAdmin(oc, "Network.operator.openshift.io/cluster", customPatchIPv4)
		} else if ipStackType == "ipv6single" {
			defer func() {
				patchResourceAsAdmin(oc, "Network.operator.openshift.io/cluster", patchIPv6original)
				err := checkOVNKState(oc)
				exutil.AssertWaitPollNoErr(err, fmt.Sprintf("OVNkube didn't trigger or rolled out successfully post oc patch"))
			}()
			patchResourceAsAdmin(oc, "Network.operator.openshift.io/cluster", customPatchIPv6)
		} else {
			defer func() {
				patchResourceAsAdmin(oc, "Network.operator.openshift.io/cluster", patchDualstackoriginal)
				err := checkOVNKState(oc)
				exutil.AssertWaitPollNoErr(err, fmt.Sprintf("OVNkube didn't trigger or rolled out successfully post oc patch"))
			}()
			patchResourceAsAdmin(oc, "Network.operator.openshift.io/cluster", customPatchDualstack)
		}
		err = checkOVNKState(oc)
		exutil.AssertWaitPollNoErr(err, fmt.Sprintf("OVNkube never trigger or rolled out successfully post oc patch"))
		//check usual svc and pod connectivities post migration which also ensures disruption doesn't last post successful rollout
		CurlPod2PodPass(oc, oc.Namespace(), pod1ns.name, oc.Namespace(), pod2ns.name)
		CurlPod2SvcPass(oc, oc.Namespace(), oc.Namespace(), pod1ns.name, "test-service-73205")
	})
})
