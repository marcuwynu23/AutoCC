# AutoCI/CD

AutoCC is a Continuous Integration/Continuous Deployment (CI/CD) tool designed to automate your application workflows. It monitors configuration files, updates repositories, and executes defined steps automatically.

## Features

- Monitors JSON configuration files in a specified directory for changes.
- Automatically pulls updates from Git repositories.
- Executes customizable steps defined in JSON configuration files.
- Supports logging for enhanced traceability.

---

## Global Configuration

The global configuration for AutoCC is stored in `/etc/autocc/settings.json` on Linux. Below is an example configuration:

```json
{
  "scriptsDirectory": "./scripts",
  "appsDirectory": "./apps",
  "ticker": 30,
  "logEnabled": true
}
```

### Configuration Fields

- **`scriptsDirectory`**: The directory where JSON scripts are located. AutoCC monitors this directory for changes.
- **`appsDirectory`**: The directory where applications and their repositories will be managed.
- **`ticker`**: The interval (in seconds) for polling remote repositories for updates.
- **`logEnabled`**: Enable or disable logging (`true` or `false`).

---

## Script Configuration

Each script defines the workflow for an application. Below is an example script:

```json
{
  "appName": "try-github-action-deploy",
  "gitRepo": "git@github.com:marcuwynu23/try-github-action-deploy.git",
  "triggerBranch": "main",
  "steps": [
    {
      "name": "read README.md",
      "cmd": "cat",
      "args": [
        "README.md"
      ]
    },
    {
      "name": "test 1",
      "cmd": "cat",
      "args": [
        "README.md"
      ]
    },
    {
      "name": "test 2",
      "cmd": "cat",
      "args": [
        "README.md"
      ]
    }
  ]
}
```

### Script Fields

- **`appName`**: The name of the application. This is used to organize application-specific files.
- **`gitRepo`**: The Git repository URL of the application.
- **`triggerBranch`**: The branch to monitor for updates.
- **`steps`**: An array of steps to execute. Each step has the following fields:
  - **`name`**: A descriptive name for the step.
  - **`cmd`**: The command to execute.
  - **`args`**: The arguments for the command.

---

## Usage

1. **Install AutoCC**: Ensure you have built and installed the `autocc` executable (see `Makefile` for details).
2. **Configure Global Settings**: Edit `/etc/autocc/settings.json` to define your global configuration.
3. **Add Scripts**: Place your script configuration files in the `scriptsDirectory` specified in `settings.json`.
4. **Start Monitoring**: Run AutoCC to begin monitoring and executing workflows automatically.

```bash
./autocc
```

---

## Example Workflow

Suppose you have the following setup:

### Global Configuration
```json
{
  "scriptsDirectory": "./scripts",
  "appsDirectory": "./apps",
  "ticker": 30,
  "logEnabled": true
}
```

### Script
```json
{
  "appName": "try-github-action-deploy",
  "gitRepo": "git@github.com:marcuwynu23/try-github-action-deploy.git",
  "triggerBranch": "main",
  "steps": [
    {
      "name": "read README.md",
      "cmd": "cat",
      "args": [
        "README.md"
      ]
    }
  ]
}
```

AutoCC will:
1. Clone the `try-github-action-deploy` repository if it does not exist.
2. Pull the latest changes from the `main` branch at the configured interval.
3. Execute the steps defined in the script, such as running the `cat README.md` command.

---

## Contributing

Feel free to contribute to AutoCC by submitting issues or pull requests to the repository.

---

## License

AutoCC is open-source software licensed under the MIT License.

---

## Support

For support, please open an issue on the GitHub repository or contact the maintainer directly.

