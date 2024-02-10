package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/jithin-kg/gomon/constants"
	"github.com/jithin-kg/gomon/internal/builder"
	"github.com/jithin-kg/gomon/internal/config"
	"github.com/jithin-kg/gomon/internal/utils"
	"github.com/jithin-kg/gomon/internal/watcher"
	"github.com/spf13/cobra"
)

func createDefaultConfig(filename string) {
	binName := utils.GetBinaryName()
	color.Yellow("binName %s", binName)

	defaultConfig := &config.Config{
		Watch:  []string{"./"}, //watch all sub directories
		Ignore: []string{"vendor/*", ".git/*", "tmp/*"},
		Build: config.BuildConfig{
			Command:   fmt.Sprintf("go build -o ./tmp/%s .", binName), // Adjust as needed for your project structure
			Directory: ".",                                            // Assumes the build is done in the current directory

		},
		Run: fmt.Sprintf("./tmp/%s", binName),
		Env: map[string]string{},
	}

	// marshal default config into JSON, with intedtation for readability
	data, err := json.MarshalIndent(defaultConfig, "", "  ")
	if err != nil {
		log.Fatalf("Failed to create default config: %v", err)
	}
	// write the json to gomon.json
	if err = os.WriteFile(filename, data, 0644); err != nil {
		log.Fatalf("Failed to write to default %s file %v\n", filename, err)
	}
	log.Printf("Created default config file %s", filename)
}

var appProcess *exec.Cmd
var mutex sync.Mutex

func runApplication(runCmd string) {
	// Mutex: A mutex is used to ensure that stopping and starting the application is thread-safe.
	// This is important because the file watcher operates in a separate goroutine
	mutex.Lock()
	defer mutex.Unlock()
	// stop currently running instances if any
	if appProcess != nil {
		if err := appProcess.Process.Kill(); err != nil {
			color.Yellow("Failed to kill running application: %v", err)
			// handle error, possibly continue attempt to run the new instance
		}
		appProcess = nil
	}
	// split the runCmd to command and args
	cmdParts := strings.Fields(runCmd)
	if len(cmdParts) == 0 {
		color.Yellow("No command specified to run the application")
		return
	}

	// prepare the command the run the built application
	appProcess = exec.Command(cmdParts[0], cmdParts[1:]...)
	appProcess.Stdout = os.Stdout
	appProcess.Stderr = os.Stderr

	// start the application
	log.Printf("Running application: %s", runCmd)

	if err := appProcess.Start(); err != nil {
		color.Red("Failed to start the application: %v", err)
	} else {
		color.Green("Application started successfully")
	}
}

func stopApplication() {
	mutex.Lock()
	defer mutex.Unlock()
	if appProcess != nil {
		// attemp to gracefully terminate the process
		if err := appProcess.Process.Signal(os.Interrupt); err != nil {
			log.Printf("Failed to send interrupt to the application: %v\n", err)
			// interrupt failed, forcefully kill the process
			if killErr := appProcess.Process.Kill(); killErr != nil {
				log.Printf("Failed to forcefully kill the application: %v\n", killErr)
			}
		}
		appProcess = nil
		log.Println("Stopped the application due to build failure")
	}
}

func buildAndRun(config *config.Config) {
	// ensure tmp directory is created
	if err := utils.EnsureTmpDirectory(); err != nil {
		log.Fatalf("Failed to ensure tmp directory exists: %v", err)
	}
	b := builder.New(config.Build.Command, config.Build.Directory)
	if err := b.Build(); err != nil {
		color.Red("Build failed %v\n", err)
		stopApplication()

	} else {
		color.Green("Build succeeded, running the application...")
		runApplication(config.Run)
	}
}

var rootCmd = &cobra.Command{
	Use:   "gomon",
	Short: "Gomon monitors your Go project files and rebuilds them on changes.",
	Run: func(cmd *cobra.Command, args []string) {

		if _, err := os.Stat(constants.ConfigFileName); os.IsNotExist(err) {
			log.Printf("Configuration file %s not found. Creating default configuration.\n", constants.ConfigFileName)
			// here we have to add the logic to create the file with actual configuration
			createDefaultConfig(constants.ConfigFileName)
		}
		appConfigs, err := config.LoadConfig(constants.ConfigFileName)
		if err != nil {
			log.Printf("Failed to load config file %s err: %v\n", constants.ConfigFileName, err)
		}
		// perform an initial build before watching for file changes
		buildAndRun(appConfigs)

		// initialise and start watcher
		w, err := watcher.New(appConfigs.Watch, appConfigs.Ignore, func() {
			// call back on file change
			color.Blue("rebuilding config : %v\n", appConfigs)
			buildAndRun(appConfigs)
		})
		if err != nil {
			log.Fatalf("Failed to create watcher %v\n", err)
		}
		w.Start()
		defer w.Close()

		// // keep the application running
		// select {}
		// main loop

		for {
			select {
			case <-w.ConfigChange:
				color.GreenString("Config file updated")
				// reload configuration
				updatedConfig, err := config.LoadConfig(constants.ConfigFileName)
				if err != nil {
					log.Printf("Failed to load config %v\n", err)
					continue
				}
				appConfigs = updatedConfig
				buildAndRun(appConfigs)

			}
		}
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}
