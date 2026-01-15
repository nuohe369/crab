package boot

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/nuohe369/crab/common/config"
	"github.com/nuohe369/crab/pkg/crypto"
)

var (
	addr        string
	secretKey   string
	serviceName string
	moduleList  string
)

// SecretKey returns the decryption key for configuration.
func SecretKey() string {
	return secretKey
}

var rootCmd = &cobra.Command{
	Use:   "crab",
	Short: "Crab framework server",
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the server",
	Long: `Start the server with various options:
  serve              Start all modules (using default port)
  serve -s admin     Start the admin service defined in config file
  serve -m admin,api Start specified modules
  serve -a :8080     Specify port`,
	Run: func(cmd *cobra.Command, args []string) {
		if serviceName != "" {
			// Start by service name
			RunService(serviceName, addr)
		} else if moduleList != "" {
			// Start by module list
			modules := strings.Split(moduleList, ",")
			for i := range modules {
				modules[i] = strings.TrimSpace(modules[i])
			}
			if addr == "" {
				addr = ":3000"
			}
			RunModules(modules, addr)
		} else {
			// Start all modules
			if addr == "" {
				addr = ":3000"
			}
			Run(addr)
		}
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered modules and service configurations",
	Run: func(cmd *cobra.Command, args []string) {
		runList()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(VersionInfo())
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate configuration file",
	Long: `Generate configuration file:
  init              Generate config.example.toml (without overwriting existing file)
  init --force      Force overwrite config.example.toml
  init --config     Generate config.toml (if it doesn't exist)`,
	Run: func(cmd *cobra.Command, args []string) {
		runInit()
	},
}

var initMenuCmd = &cobra.Command{
	Use:   "init-menu",
	Short: "Generate menu configuration file menu.toml",
	Run: func(cmd *cobra.Command, args []string) {
		runInitMenu()
	},
}

var encryptCmd = &cobra.Command{
	Use:   "encrypt",
	Short: "Encrypt configuration values",
	Run: func(cmd *cobra.Command, args []string) {
		runEncrypt()
	},
}

var encryptValue string
var initForce bool
var initConfig bool

func init() {
	rootCmd.PersistentFlags().StringVarP(&addr, "addr", "a", "", "Listen address")
	rootCmd.PersistentFlags().StringVarP(&secretKey, "key", "k", "", "Configuration decryption key")

	serveCmd.Flags().StringVarP(&serviceName, "service", "s", "", "Service name (from config file)")
	serveCmd.Flags().StringVarP(&moduleList, "modules", "m", "", "Module list (comma-separated)")

	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "Force overwrite")
	initCmd.Flags().BoolVarP(&initConfig, "config", "c", false, "Generate config.toml")

	encryptCmd.Flags().StringVarP(&encryptValue, "value", "v", "", "Value to encrypt")
	encryptCmd.MarkFlagRequired("value")

	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(initMenuCmd)
	rootCmd.AddCommand(encryptCmd)
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// runList lists all registered modules and service configurations.
func runList() {
	// Load configuration
	if secretKey != "" {
		config.SetDecryptKey(secretKey)
	}
	config.MustLoad("config.toml")

	fmt.Println("Registered modules:")
	for _, name := range GetAllModuleNames() {
		fmt.Printf("  - %s\n", name)
	}

	fmt.Println("\nService configurations:")
	for _, svc := range config.GetServices() {
		fmt.Printf("  [%s] %s -> %v\n", svc.Name, svc.Addr, svc.Modules)
	}
}

// runInit generates configuration file.
func runInit() {
	// Determine target file
	configPath := "config.example.toml"
	if initConfig {
		configPath = "config.toml"
	}

	// Check if file exists
	if _, err := os.Stat(configPath); err == nil && !initForce {
		fmt.Printf("%s already exists, use --force to overwrite\n", configPath)
		return
	}

	content := `# Crab Framework Configuration File

[app]
name = "server"
version = "0.1.0"
env = "dev"  # dev: development, prod: production

[server]
addr = ":3000"

# ==================== Database Configuration ====================
[database]
host = "localhost"
port = 5432
user = "server"
password = "server"
db_name = "server"

# ==================== Redis Configuration ====================
[redis]
# Standalone mode
addr = "localhost:6379"
password = ""
db = 0
# Cluster mode (automatically switches when configured, addr and db will be ignored)
# cluster = "host1:6379,host2:6379,host3:6379"

# ==================== Message Queue Configuration ====================
[mq]
driver = ""  # redis or rabbitmq, leave empty to disable

[mq.redis]
addr = "localhost:6379"
password = ""
db = 0
max_len = 10000  # Stream max length, 0 means unlimited
# cluster = "host1:6379,host2:6379,host3:6379"

[mq.rabbitmq]
url = "amqp://guest:guest@localhost:5672/"

# ==================== JWT Configuration ====================
[jwt]
secret = "your-jwt-secret-change-me"
expire = "24h"

# ==================== Tracing Configuration ====================
[trace]
service_name = "server"
endpoint = ""  # Leave empty to disable, e.g.: localhost:4318
insecure = true

# ==================== Metrics Configuration ====================
[metrics]
enabled = false
path = "/metrics"

# ==================== Storage Configuration ====================
[storage]
driver = ""  # local, oss, s3, leave empty to disable

[storage.local]
root = "./uploads"
base_url = "/uploads"  # For frontend-backend separation, configure full URL, e.g.: https://api.example.com/uploads

[storage.oss]
endpoint = ""
access_key_id = ""
access_key_secret = ""
bucket = ""
base_url = ""  # CDN domain

[storage.s3]
region = "us-east-1"
access_key_id = ""
secret_access_key = ""
bucket = ""
endpoint = ""  # Custom endpoint for MinIO, etc.
base_url = ""

# ==================== Service Configuration ====================
# Define different service combinations, start with serve -s <name>

[[services]]
name = "test"
addr = ":3001"
modules = ["testapi", "ws"]
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		fmt.Printf("Failed to generate configuration file: %v\n", err)
		return
	}
	fmt.Printf("%s generated successfully\n", configPath)

	if !initConfig {
		fmt.Println("Tip: Copy to config.toml and modify the configuration before use")
	}
}

// runEncrypt encrypts configuration values.
func runEncrypt() {
	if secretKey == "" {
		fmt.Println("Error: Please specify a key with -k")
		os.Exit(1)
	}

	encrypted, err := crypto.Encrypt(encryptValue, secretKey)
	if err != nil {
		fmt.Printf("Encryption failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Encrypted result:")
	fmt.Println(encrypted)
}

// runInitMenu generates menu configuration file (not needed in framework version).
func runInitMenu() {
	fmt.Println("Menu configuration is not needed in framework version")
}
