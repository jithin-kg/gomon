gomon/
│
├── cmd/
│   └── gomon/
│       └── main.go  # Cobra application entry point
│
├── internal/
│   ├── builder/
│   │   └── builder.go  # Logic for building the Go project
│   │
│   ├── config/
│   │   └── config.go  # Configuration loading and parsing
│   │
│   ├── watcher/
│   │   └── watcher.go  # Filesystem watching and event handling
│   │
│   └── utils/
│       └── utils.go   # Utility functions, including getBinaryName and directory management
│
└── go.mod
└── go.sum
