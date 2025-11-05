package util

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CLI provides a minimal interface for test operations
type CLI struct {
	namespace string
}

// Namespace returns the test namespace
func (c *CLI) Namespace() string {
	return c.namespace
}

// KubeFramework returns a minimal framework interface
func (c *CLI) KubeFramework() *Framework {
	return &Framework{}
}

// AsAdmin returns the CLI for admin operations
func (c *CLI) AsAdmin() *AdminCLI {
	return &AdminCLI{cli: c}
}

// WithoutNamespace returns the CLI without namespace context
func (c *CLI) WithoutNamespace() *CLI {
	return c
}

// AdminCLI provides admin CLI operations
type AdminCLI struct {
	cli *CLI
}

// WithoutNamespace returns the admin CLI without namespace context
func (a *AdminCLI) WithoutNamespace() *AdminCLI {
	return a
}

// Run returns a command runner
func (a *AdminCLI) Run(command string) *CommandRunner {
	return &CommandRunner{command: command}
}

// CommandRunner provides command execution
type CommandRunner struct {
	command string
	args    []string
}

// Args sets command arguments
func (r *CommandRunner) Args(args ...string) *CommandRunner {
	r.args = args
	return r
}

// Output executes and returns output
func (r *CommandRunner) Output() (string, error) {
	return "", fmt.Errorf("not implemented")
}

// Framework provides minimal test framework functionality
type Framework struct{}

// ClientSet returns a function that returns a nil clientset (to be implemented)
func (f *Framework) ClientSet() func() kubernetes.Interface {
	return func() kubernetes.Interface {
		return nil
	}
}

// By logs a test step
func By(text string) {
	fmt.Printf("STEP: %s\n", text)
}

// Logf logs a formatted message
func Logf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// RunHostCmd runs a command in a pod
func RunHostCmd(namespace, podName, cmd string) (string, error) {
	return "", fmt.Errorf("not implemented: RunHostCmd")
}

// RunHostCmdWithRetries runs a command in a pod with retries
func RunHostCmdWithRetries(namespace, podName, cmd string, interval, timeout time.Duration) (string, error) {
	return "", fmt.Errorf("not implemented: RunHostCmdWithRetries")
}

// FixturePath returns a path to test fixtures
func FixturePath(elem ...string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "testdata", filepath.Join(elem...))
}

// AssertWaitPollNoErr checks for polling errors
func AssertWaitPollNoErr(err error, msg string) {
	if err != nil {
		panic(fmt.Sprintf("%s: %v", msg, err))
	}
}

// GetReadySchedulableNodes returns ready and schedulable nodes
func GetReadySchedulableNodes(ctx context.Context, c kubernetes.Interface) (*corev1.NodeList, error) {
	if c == nil {
		return &corev1.NodeList{}, fmt.Errorf("clientset is nil")
	}
	nodes, err := c.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	readyNodes := &corev1.NodeList{}
	for _, node := range nodes.Items {
		if isNodeReady(&node) && isNodeSchedulable(&node) {
			readyNodes.Items = append(readyNodes.Items, node)
		}
	}
	return readyNodes, nil
}

func isNodeReady(node *corev1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func isNodeSchedulable(node *corev1.Node) bool {
	return !node.Spec.Unschedulable
}

