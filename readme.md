<p align="center">
  <img src="./gomon.jpg" alt="Gomon Logo" width="200" height="200"/>
</p>

# Gomon

Gomon is a live reloader for Go applications, inspired by Nodemon. It monitors changes in your Go project files and automatically rebuilds and restarts your application, making development faster and more interactive.

## Features

- **Automatic Rebuilding:** Automatically rebuilds your Go application whenever source files change.
- **Customizable Watch List:** Configure directories or files to watch for changes.
- **Support for Various Notifications:** Get desktop notifications on build success or failure (future feature).
- **Cross-Platform:** Works on Windows, macOS, and Linux.

## Getting Started

### Prerequisites

- Go 1.15 or later

### Installation

Install Gomon by running:

```sh
go install github.com/jithin-kg/gomon@latest
```

## Usage

To start using Gomon, navigate to your project directory and run:

```sh
gomon
```

This command will start monitoring your Go files for changes and reload your application as needed.

## Configuration

Gomon can be configured via a `gomon.json` file in your project root. Here's an example configuration:

```json
{
  "watch": ["./..."],
  "ignore": ["vendor/*", ".git/*", "tmp/*"],
  "build": {
    "command": "go build -o myapp .",
    "directory": "."
  },
  "run": "./myapp"
}
```

For a detailed explanation of each configuration option, please visit the Configuration section.

## Contributing

We welcome contributions from the community, whether it's in the form of bug reports, feature requests, or pull requests.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -am 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a pull request

Please read `CONTRIBUTING.md` for details on our code of conduct and the process for submitting pull requests.

## License

Gomon is open-source software licensed under the MIT License - see the `LICENSE` file for details.

## Acknowledgments

- Thanks to the creators of Nodemon for the inspiration.
- Thanks to all contributors who help in improving Gomon.
