package main

import (
    "github.com/fsnotify/fsnotify"
    "github.com/will-wright-eng/discoverer/internal/service"
    "log"
    "os"
    "time"
	"os/signal"
	"syscall"
)

func main() {
    log.SetFlags(log.LstdFlags | log.Lshortfile)

    servicesDir := os.Getenv("SERVICES_DIR")
    nginxConfPath := os.Getenv("NGINX_CONF_PATH")
    nginxPidPath := os.Getenv("NGINX_PID_PATH")

    log.Printf("SERVICES_DIR=%s", servicesDir)
    log.Printf("NGINX_CONF_PATH=%s", nginxConfPath)
    log.Printf("NGINX_PID_PATH=%s", nginxPidPath)

    if servicesDir == "" {
        servicesDir = "./services"
    }
    if nginxConfPath == "" {
        nginxConfPath = "/etc/nginx/nginx.conf"
    }
    if nginxPidPath == "" {
        nginxPidPath = "/var/run/nginx/nginx.pid"
    }

    if _, err := os.Stat(servicesDir); os.IsNotExist(err) {
        log.Printf("Services directory does not exist: %s", servicesDir)
        if err := os.MkdirAll(servicesDir, 0755); err != nil {
            log.Fatalf("Failed to create services directory: %v", err)
        }
        log.Printf("Created services directory: %s", servicesDir)
    }

    log.Println("Waiting for nginx to start...")
    timeout := time.After(30 * time.Second)
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            if _, err := os.Stat(nginxPidPath); err == nil {
                log.Println("Nginx is ready")
                goto nginxReady
            }
        case <-timeout:
            log.Fatalf("Timeout waiting for nginx to start")
        }
    }
nginxReady:

    manager := service.NewManager(servicesDir, nginxConfPath, nginxPidPath)
    if err := manager.LoadServices(); err != nil {
        log.Printf("Warning: Failed to load initial services: %v", err)
    }

    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatalf("Failed to create watcher: %v", err)
    }
    defer watcher.Close()

    if err := watcher.Add(servicesDir); err != nil {
        log.Fatalf("Failed to watch services directory: %v", err)
    }

    log.Printf("Watching directory: %s", servicesDir)

    done := make(chan bool)

    go func() {
        for {
            select {
            case event, ok := <-watcher.Events:
                if !ok {
                    done <- true
                    return
                }
                if event.Op&fsnotify.Write == fsnotify.Write ||
                   event.Op&fsnotify.Create == fsnotify.Create ||
                   event.Op&fsnotify.Remove == fsnotify.Remove {
                    log.Printf("Detected change in %s", event.Name)
                    if err := manager.LoadServices(); err != nil {
                        log.Printf("Error reloading services: %v", err)
                    }
                }
            case err, ok := <-watcher.Errors:
                if !ok {
                    done <- true
                    return
                }
                log.Printf("Watcher error: %v", err)
            }
        }
    }()

    go func() {
        sigchan := make(chan os.Signal, 1)
        signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
        <-sigchan
        log.Println("Received shutdown signal")
        done <- true
    }()

    log.Println("Service discoverer is running...")
    <-done
    log.Println("Shutting down service discoverer")
}
