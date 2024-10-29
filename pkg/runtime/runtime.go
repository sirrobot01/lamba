package runtime

import (
	"context"
	"fmt"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/sirrobot01/lamba/pkg/event"
	"github.com/sirrobot01/lamba/pkg/function"
	"os"
	"os/exec"
	"sync"
	"time"
)

// Global containerd client
var (
	containerdClient *containerd.Client
	defaultTimeout   = 20 * time.Second
)

func GetContainerdClient() *containerd.Client {
	return containerdClient
}

type Runtime interface {
	Init(fn *function.Function) error
	Shutdown(fn *function.Function) error
	GetCmd(event *event.Event, fn *function.Function) []string
	GetImage() string
}

type Manager struct {
	runtimes map[string]Runtime
	mu       sync.Mutex
}

func NewManager() (*Manager, error) {
	// Print diagnostic information
	if err := diagnoseContainerdSetup(); err != nil {
		return nil, err
	}
	var err error
	containerdClient, err = initContainerdClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize containerd client: %v", err)
	}

	// Test listing images to verify connection
	ctx := namespaces.WithNamespace(context.Background(), "moby")
	_, err = containerdClient.ImageService().List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %v", err)
	}

	return &Manager{
		runtimes: make(map[string]Runtime),
	}, nil
}

func getContainerdSocket() string {
	return "/tmp/containerd.sock"
}

func initContainerdClient() (*containerd.Client, error) {
	socketPath := getContainerdSocket()
	if socketPath == "" {
		return nil, fmt.Errorf("containerd socket not found")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connection options
	opts := []containerd.ClientOpt{
		containerd.WithDefaultNamespace("default"),
		containerd.WithTimeout(30 * time.Second),
		containerd.WithDefaultRuntime("io.containerd.runc.v2"),
	}

	// Create client
	client, err := containerd.New(socketPath, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create containerd client: %v", err)
	}

	// Test connection
	if _, err := client.Version(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect to containerd: %v", err)
	}

	return client, nil
}

func checkColimaStatus() error {
	cmd := exec.Command("colima", "status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("colima status check failed: %v\nOutput: %s", err, output)
	}
	fmt.Printf("Colima status: %s\n", output)
	return nil
}

func diagnoseContainerdSetup() error {
	fmt.Println("Diagnosing containerd setup...")

	// Check Colima status
	if err := checkColimaStatus(); err != nil {
		return err
	}

	socketPath := getContainerdSocket()
	if socketPath == "" {
		return fmt.Errorf("containerd socket path could not be determined")
	}

	fmt.Printf("Found containerd socket at: %s\n", socketPath)

	fileInfo, err := os.Stat(socketPath)
	if err != nil {
		return fmt.Errorf("error accessing socket: %v", err)
	}

	fmt.Printf("Socket permissions: %v\n", fileInfo.Mode())

	fmt.Printf("colima nerdctl test  --address unix://%s ps\n", socketPath)

	// Try nerdctl command
	cmd := exec.Command("nerdctl", "--address", fmt.Sprintf("unix://%s", socketPath), "ps")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("nerdctl test failed: %v\nOutput: %s\n", err, output)
	} else {
		fmt.Printf("nerdctl test succeeded: %s\n", output)
	}

	return nil
}

func (m *Manager) Register(runtimes map[string]Runtime) error {
	var wg sync.WaitGroup
	errorChan := make(chan error, len(runtimes))

	for name, rtn := range runtimes {
		wg.Add(1)
		go func(name string, rtn Runtime) {
			defer wg.Done()
			if err := rtn.Init(nil); err != nil {
				errorChan <- err
			}
			m.mu.Lock()
			m.runtimes[name] = rtn
			m.mu.Unlock()
		}(name, rtn)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	// Close the channel after all goroutines are done
	close(errorChan)

	// Check for any errors
	for err := range errorChan {
		if err != nil {
			return err // Return the first error encountered
		}
	}

	return nil
}

func (m *Manager) Get(name string) (Runtime, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rtn, exists := m.runtimes[name]
	return rtn, exists
}

func (m *Manager) List() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	runtimes := make([]string, 0, len(m.runtimes))
	for rtn := range m.runtimes {
		runtimes = append(runtimes, rtn)
	}
	return runtimes
}

func (m *Manager) Shutdown(fn *function.Function) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, rtn := range m.runtimes {
		err := rtn.Shutdown(fn)
		if err != nil {
			return
		}
	}
}

// Close closes the containerd client connection
func (m *Manager) Close() error {
	if containerdClient != nil {
		return containerdClient.Close()
	}
	return nil
}
