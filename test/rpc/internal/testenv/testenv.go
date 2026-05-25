package testenv

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	LocalEnv             = "GO_MALL_TEST_LOCAL"
	NamespaceEnv         = "GO_MALL_TEST_NAMESPACE"
	ServiceEndpointsEnv  = "GO_MALL_TEST_SERVICE_ENDPOINTS"
	TimeoutEnv           = "GO_MALL_TEST_TIMEOUT"
	RunIDEnv             = "GO_MALL_TEST_RUN_ID"
	DefaultNamespace     = "go-mall"
	DefaultTimeout       = 30 * time.Second
	DefaultServiceDomain = "svc.cluster.local"
)

var (
	endpointsOnce sync.Once
	endpoints     map[string]string
	runIDOnce     sync.Once
	runIDValue    string
)

func Namespace() string {
	if v := strings.TrimSpace(os.Getenv(NamespaceEnv)); v != "" {
		return v
	}
	return DefaultNamespace
}

func Timeout() time.Duration {
	if v := strings.TrimSpace(os.Getenv(TimeoutEnv)); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return DefaultTimeout
}

func RunID() string {
	runIDOnce.Do(func() {
		if v := strings.TrimSpace(os.Getenv(RunIDEnv)); v != "" {
			runIDValue = v
			return
		}
		runIDValue = uuid.NewString()
	})
	return runIDValue
}

func ServiceAddr(service string, port int) string {
	if v := strings.TrimSpace(serviceEndpointOverrides()[service]); v != "" {
		return v
	}
	if strings.Contains(service, ":") {
		return service
	}
	if LocalMode() {
		return fmt.Sprintf("127.0.0.1:%d", port)
	}
	return fmt.Sprintf("%s-rpc.%s.%s:%d", service, Namespace(), DefaultServiceDomain, port)
}

func UniqueName(prefix string) string {
	return fmt.Sprintf("%s-%s-%s", prefix, RunID(), uuid.NewString()[:8])
}

func serviceEndpointOverrides() map[string]string {
	endpointsOnce.Do(func() {
		endpoints = make(map[string]string)
		raw := strings.TrimSpace(os.Getenv(ServiceEndpointsEnv))
		if raw == "" {
			return
		}
		for _, item := range strings.Split(raw, ",") {
			parts := strings.SplitN(strings.TrimSpace(item), "=", 2)
			if len(parts) != 2 {
				continue
			}
			name := strings.TrimSpace(parts[0])
			addr := strings.TrimSpace(parts[1])
			if name != "" && addr != "" {
				endpoints[name] = addr
			}
		}
	})
	return endpoints
}

func LocalMode() bool {
	return strings.TrimSpace(os.Getenv(LocalEnv)) == "1"
}
