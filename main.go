package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"runtime"

	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
)

type Step struct {
	Name string   `json:"name"`
	Cmd  string   `json:"cmd"`
	Args []string `json:"args"`
}

type Config struct {
	AppName       string `json:"appName"`
	GitRepo       string `json:"gitRepo"`
	TriggerBranch string `json:"triggerBranch"`
	Steps         []Step `json:"steps"`
}

type GlobalConfig struct {
	ScriptsDirectory string `json:"scriptsDirectory"`
	AppsDirectory    string `json:"appsDirectory"`
	Ticker           int    `json:"ticker"`
	LogEnabled       bool   `json:"logEnabled"`
}

var globalConfig GlobalConfig

func main() {
	// Load global settings
	if err := loadGlobalConfig("settings.json"); err != nil {
		log.Fatalf("%s", color.RedString("Error loading global settings: %v", err))
	}

	// Set the number of CPU cores Go can use
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Use the configured directories from global settings
	scriptsDir := globalConfig.ScriptsDirectory
	appsDir := globalConfig.AppsDirectory

	if scriptsDir == "" {
		scriptsDir = "./scripts" // Default if not provided
	}

	if appsDir == "" {
		appsDir = "./apps" // Default if not provided
	}

	// Start monitoring the directory
	watchConfig(scriptsDir, appsDir)
}

func loadGlobalConfig(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("unable to open settings file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&globalConfig); err != nil {
		return fmt.Errorf("unable to decode settings file: %v", err)
	}

	return nil
}

func watchConfig(scriptsDir, appsDir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("%s", color.RedString("Error creating watcher: %v", err))
	}
	defer watcher.Close()

	if err := watcher.Add(scriptsDir); err != nil {
		log.Fatalf("%s", color.RedString("Failed to watch directory %s: %v", scriptsDir, err))
	}

	log.Printf("%s", color.GreenString("Daemon started. Monitoring %s for changes...", scriptsDir))

	// Process all configuration files initially
	processScripts(scriptsDir, appsDir)

	// Poll the remote repositories every specified interval
	ticker := time.NewTicker(time.Duration(globalConfig.Ticker) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Printf("%s", color.CyanString("Polling remote repositories for updates..."))
			pollRemoteRepositories(scriptsDir, appsDir)

		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				if filepath.Ext(event.Name) == ".json" {
					log.Printf("%s", color.YellowString("Detected changes in %s. Reloading configuration...", event.Name))
					processScripts(scriptsDir, appsDir)
				}
			}

		case err := <-watcher.Errors:
			log.Printf("%s", color.RedString("Watcher error: %v", err))
		}
	}
}

func pollRemoteRepositories(scriptsDir, appsDir string) {
	files, err := os.ReadDir(scriptsDir)
	if err != nil {
		log.Printf("%s", color.RedString("Failed to read scripts directory %s: %v", scriptsDir, err))
		return
	}

	var wg sync.WaitGroup

	// Iterate over each JSON configuration file in the scripts folder
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			wg.Add(1)
			go func(fileName string) {
				defer wg.Done()
				processConfigFile(scriptsDir, appsDir, fileName)
			}(file.Name())
		}
	}

	wg.Wait()
}

func processScripts(scriptsDir, appsDir string) {
	files, err := os.ReadDir(scriptsDir)
	if err != nil {
		log.Printf("%s", color.RedString("Failed to read scripts directory %s: %v", scriptsDir, err))
		return
	}

	var wg sync.WaitGroup

	// Iterate over each JSON configuration file in the scripts folder
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			wg.Add(1)
			go func(fileName string) {
				defer wg.Done()
				processConfigFile(scriptsDir, appsDir, fileName)
			}(file.Name())
		}
	}

	wg.Wait()
}

func processConfigFile(scriptsDir, appsDir, fileName string) {
	configFile := filepath.Join(scriptsDir, fileName)
	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Printf("%s", color.RedString("Failed to read config file %s: %v", configFile, err))
		return
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		log.Printf("%s", color.RedString("Failed to parse JSON config in file %s: %v", configFile, err))
		return
	}

	log.Printf("%s", color.GreenString("[%s] Processing configuration from file %s...", config.AppName, configFile))

	appDir := filepath.Join(appsDir, config.AppName)
	if err := os.MkdirAll(appDir, os.ModePerm); err != nil {
		log.Printf("%s", color.RedString("[%s] Failed to create app directory %s: %v", config.AppName, appDir, err))
		return
	}

	repoDir := filepath.Join(appDir, "repo")
	cloneOrPull(repoDir, config)

	// Execute steps only if repository was cloned or branch changed
	if shouldExecuteSteps(repoDir, config.TriggerBranch) {
		executeSteps(repoDir, config.Steps, config.AppName)
	}
}

// Clone or pull the repository
func cloneOrPull(repoDir string, config Config) {
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		log.Printf("%s", color.YellowString("[%s] Cloning repository %s...", config.AppName, config.GitRepo))
		cmd := exec.Command("git", "clone", config.GitRepo, repoDir)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("%s", color.RedString("[%s] Failed to clone repository: %v", config.AppName, err))
			log.Printf("%s", color.RedString("Git output: %s", string(output)))
			return
		}
	} else {
		// Fetch the latest changes and checkout the branch
		log.Printf("%s", color.YellowString("[%s] Pulling latest changes from the repository...", config.AppName))
		cmd := exec.Command("git", "fetch")
		cmd.Dir = repoDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("%s", color.RedString("[%s] Failed to fetch changes from the repository: %v", config.AppName, err))
			log.Printf("%s", color.RedString("Git output: %s", string(output)))
			return
		}

		// Checkout the trigger branch
		log.Printf("%s", color.GreenString("[%s] Checking out branch: %s", config.AppName, config.TriggerBranch))
		cmd = exec.Command("git", "checkout", config.TriggerBranch)
		cmd.Dir = repoDir
		output, err = cmd.CombinedOutput()
		if err != nil {
			log.Printf("%s", color.RedString("[%s] Failed to checkout branch: %v", config.AppName, err))
			log.Printf("%s", color.RedString("Git output: %s", string(output)))
			return
		}

		// Pull the latest changes
		cmd = exec.Command("git", "pull", "origin", config.TriggerBranch)
		cmd.Dir = repoDir
		output, err = cmd.CombinedOutput()
		if err != nil {
			log.Printf("%s", color.RedString("[%s] Failed to pull latest changes: %v", config.AppName, err))
			log.Printf("%s", color.RedString("Git output: %s", string(output)))
			return
		}
	}
}

// Check if the steps should be executed
func shouldExecuteSteps(repoDir, triggerBranch string) bool {
	// Check if the repository is cloned
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		log.Printf("%s", color.YellowString("Repository is not cloned. Steps will be executed as the repository needs to be initialized."))
		return true
	}

	// Check if the remote repository 'origin' is set
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = repoDir
	remoteURL, err := cmd.CombinedOutput()
	if err != nil || len(remoteURL) == 0 {
		log.Printf("%s", color.RedString("No remote repository 'origin' found. Please ensure the repository is correctly cloned and has the remote set up."))
		return false
	}

	// Fetch the latest changes from the remote repository
	log.Printf("%s", color.CyanString("Fetching the latest changes from the remote repository..."))
	cmd = exec.Command("git", "fetch")
	cmd.Dir = repoDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("%s", color.RedString("Failed to fetch from remote repository: %v", err))
		log.Printf("%s", color.RedString("Git output: %s", string(output)))
		return false
	}

	// Get the latest commit hash from the remote for the trigger branch
	cmd = exec.Command("git", "ls-remote", "origin", triggerBranch)
	cmd.Dir = repoDir
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Printf("%s", color.RedString("Failed to fetch remote commit hash for branch %s: %v", triggerBranch, err))
		log.Printf("%s", color.RedString("Git output: %s", string(output)))
		return false
	}

	// Extract the commit hash from the output
	latestCommitHash := string(output)
	latestCommitHash = latestCommitHash[:len(latestCommitHash)-1] // Remove newline

	// Read the saved last commit hash from the file
	lastCommitFile := filepath.Join(repoDir, ".lastcommit")
	lastCommitHash, err := os.ReadFile(lastCommitFile)

	// If no previous commit hash is saved or the hash doesn't match the latest commit, proceed with execution
	if err != nil || string(lastCommitHash) != latestCommitHash {
		log.Printf("%s", color.YellowString("Repository has new commit or no previous commit hash found. Steps will be executed."))
		// Save the new commit hash for future reference
		err = os.WriteFile(lastCommitFile, []byte(latestCommitHash), 0644)
		if err != nil {
			log.Printf("%s", color.RedString("Failed to save new commit hash: %v", err))
			return false
		}
		return true
	}

	// If the commit is the same, don't execute the steps
	log.Printf("%s", color.GreenString("Repository is up-to-date with the latest commit. Steps will not be executed."))
	return false
}

func executeSteps(repoDir string, steps []Step, appName string) {
	var wg sync.WaitGroup

	for _, step := range steps {
		wg.Add(1)
		go func(step Step) {
			defer wg.Done()
			log.Printf("%s", color.BlueString("[%s] Executing step: %s", appName, step.Name))
			cmd := exec.Command(step.Cmd, step.Args...)
			cmd.Dir = repoDir
			output, err := cmd.CombinedOutput()
			if err != nil {
				log.Printf("%s", color.RedString("[%s] Step %s failed: %v", appName, step.Name, err))
				log.Printf("%s", color.RedString("Step output: %s", string(output)))
				return
			}
			log.Printf("%s", color.GreenString("[%s] Output of step %s: %s", appName, step.Name, string(output)))
		}(step)
	}

	wg.Wait()
}
