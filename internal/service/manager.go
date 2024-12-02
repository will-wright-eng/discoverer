package service

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "path/filepath"
    "syscall"
    "text/template"
    "time"
)

const nginxTemplate = `
pid /var/run/nginx/nginx.pid;

events {
    worker_connections 1024;
}

http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;

    log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
                      '$status $body_bytes_sent "$http_referer" '
                      '"$http_user_agent" "$http_x_forwarded_for"';

    access_log  /var/log/nginx/access.log  main;
    error_log   /var/log/nginx/error.log;

    sendfile        on;
    keepalive_timeout  65;

    resolver 127.0.0.11 valid=30s;

    {{range .Services}}
    upstream {{.Name}} {
        server {{.Host}}:{{.Port}};
    }
    {{end}}

    server {
        listen 80;
        server_name localhost;

        {{range .Services}}
        location {{.Path}} {
            rewrite ^{{.Path}}$ {{.Path}}/ permanent;
        }

        location {{.Path}}/ {
            proxy_pass http://{{.Name}}/;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
        {{end}}
    }
}
`

type Manager struct {
    servicesDir    string
    nginxConfPath  string
    nginxPidPath   string
    services       ServiceConfig
}

func NewManager(servicesDir, nginxConfPath, nginxPidPath string) *Manager {
    return &Manager{
        servicesDir:    servicesDir,
        nginxConfPath:  nginxConfPath,
        nginxPidPath:   nginxPidPath,
        services:       ServiceConfig{},
    }
}

func (m *Manager) LoadServices() error {
    files, err := filepath.Glob(filepath.Join(m.servicesDir, "*.json"))
    if err != nil {
        return fmt.Errorf("failed to list service files: %w", err)
    }

    // Reset services slice
    m.services.Services = []Service{}

    // If no files found, just generate config with empty services
    if len(files) == 0 {
        log.Println("No service files found")
        return m.generateConfig()
    }

    var services []Service
    for _, file := range files {
        data, err := ioutil.ReadFile(file)
        if err != nil {
            log.Printf("Warning: failed to read service file %s: %v", file, err)
            continue
        }

        // Trim any whitespace or newlines
        data = bytes.TrimSpace(data)

        var service Service
        if err := json.Unmarshal(data, &service); err != nil {
            log.Printf("Warning: failed to parse service file %s: %v\nContent: %s", file, err, string(data))
            continue
        }

        // Validate service data
        if service.Name == "" || service.Host == "" || service.Port == 0 || service.Path == "" || service.Protocol == "" {
            log.Printf("Warning: invalid service configuration in %s: missing required fields", file)
            continue
        }

        log.Printf("Loaded service: %+v", service)
        services = append(services, service)
    }

    m.services.Services = services
    log.Printf("Total services loaded: %d", len(services))

    err = m.generateConfig()
    if err != nil {
        return fmt.Errorf("failed to generate nginx config: %w", err)
    }

    log.Printf("Successfully generated nginx config with %d services", len(services))
    return nil
}

func (m *Manager) generateConfig() error {
    // Parse template
    tmpl, err := template.New("nginx").Parse(nginxTemplate)
    if err != nil {
        return fmt.Errorf("failed to parse nginx template: %w", err)
    }

    // Create nginx config file
    f, err := os.Create(m.nginxConfPath)
    if err != nil {
        return fmt.Errorf("failed to create nginx config file: %w", err)
    }
    defer f.Close()

    // Execute template
    if err := tmpl.Execute(f, m.services); err != nil {
        return fmt.Errorf("failed to execute template: %w", err)
    }

    return m.reloadNginx()
}

func (m *Manager) reloadNginx() error {
    // Read nginx PID
    pidBytes, err := ioutil.ReadFile(m.nginxPidPath)
    if err != nil {
        return fmt.Errorf("failed to read nginx pid file: %w", err)
    }

    var pid int
    _, err = fmt.Sscanf(string(pidBytes), "%d", &pid)
    if err != nil {
        return fmt.Errorf("failed to parse nginx pid: %w", err)
    }

    // Send SIGHUP signal to nginx
    process, err := os.FindProcess(pid)
    if err != nil {
        return fmt.Errorf("failed to find nginx process: %w", err)
    }

    err = process.Signal(syscall.SIGHUP)
    if err != nil {
        return fmt.Errorf("failed to send reload signal to nginx: %w", err)
    }

    log.Printf("Sent reload signal to nginx process %d", pid)

    // Wait a moment to ensure nginx has reloaded
    time.Sleep(100 * time.Millisecond)

    return nil
}

func (m *Manager) GetServices() []Service {
    return m.services.Services
}
