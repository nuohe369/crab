package boot

import (
	"fmt"
	"os"
	"strings"

	"github.com/nuohe369/crab/common/config"
	"github.com/nuohe369/crab/pkg/crypto"
	"github.com/spf13/cobra"
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

var depsCmd = &cobra.Command{
	Use:   "deps",
	Short: "Show module dependencies",
	Long:  `Show database dependencies for all registered modules`,
	Run: func(cmd *cobra.Command, args []string) {
		runDeps()
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
  init         Generate config.toml with minimal required configuration
  init --full  Generate config.example.toml with all available options`,
	Run: func(cmd *cobra.Command, args []string) {
		runInit()
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
var initFull bool

func init() {
	rootCmd.PersistentFlags().StringVarP(&addr, "addr", "a", "", "Listen address")
	rootCmd.PersistentFlags().StringVarP(&secretKey, "key", "k", "", "Configuration decryption key")

	serveCmd.Flags().StringVarP(&serviceName, "service", "s", "", "Service name (from config file)")
	serveCmd.Flags().StringVarP(&moduleList, "modules", "m", "", "Module list (comma-separated)")

	initCmd.Flags().BoolVarP(&initFull, "full", "f", false, "Generate full configuration with all options")

	encryptCmd.Flags().StringVarP(&encryptValue, "value", "v", "", "Value to encrypt")
	encryptCmd.MarkFlagRequired("value")

	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(depsCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
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

// runDeps shows module dependencies
// runDeps 显示模块依赖
func runDeps() {
	// Load configuration
	if secretKey != "" {
		config.SetDecryptKey(secretKey)
	}
	config.MustLoad("config.toml")

	// Initialize infrastructure to get database connections
	// 初始化基础设施以获取数据库连接
	initBase()

	// Print module dependencies
	// 打印模块依赖
	PrintModuleDependencies(modules)
}

// runInit generates configuration file.
func runInit() {
	var configPath string
	var content string

	if initFull {
		// Generate full configuration example
		configPath = "config.example.toml"
		content = getFullConfig()
	} else {
		// Generate minimal required configuration
		configPath = "config.toml"
		content = getMinimalConfig()
	}

	// Check if file exists
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("%s already exists, skipping generation\n", configPath)
		fmt.Printf("Tip: Delete the file first if you want to regenerate it\n")
		return
	}

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		fmt.Printf("Failed to generate configuration file: %v\n", err)
		return
	}
	fmt.Printf("%s generated successfully\n", configPath)

	if !initFull {
		fmt.Println("\nTip: Run 'init --full' to see all available configuration options")
	}
}

// getMinimalConfig returns minimal required configuration
func getMinimalConfig() string {
	return `# Crab Framework Configuration File
# This is the minimal required configuration to get started

[app]
name = "crab"
version = "0.1.0"
env = "dev"  # dev: development, prod: production
strict_dependency_check = true  # Strict dependency checking (true in prod by default)

[server]
addr = ":3000"

# ==================== Snowflake ID Generator ====================
[snowflake]
machine_id = 1  # Machine ID (0-1023), must be unique in distributed environment

# ==================== Database Configuration (Required) ====================
[database.default]
host = "localhost"
port = 5432
user = "crab"
password = "crab"
db_name = "crab"
auto_migrate = true   # Auto migrate database schema
show_sql = false      # Show SQL logs

# ==================== Redis Configuration (Required) ====================
[redis.default]
addr = "localhost:6379"
password = ""
db = 0

# ==================== Service Configuration ====================
# Define different service combinations, start with: serve -s <name>

[[services]]
name = "all"
addr = ":3000"
modules = ["testapi", "ws"]
`
}

// getFullConfig returns full configuration with all options
func getFullConfig() string {
	return `# Crab Framework Configuration File
# This file contains all available configuration options

[app]
name = "crab"
version = "0.1.0"
env = "dev"  # dev: development, prod: production
strict_dependency_check = true  # Strict dependency checking (default: true in prod, false in dev)
                                # When enabled, modules with missing database dependencies will not start

[server]
addr = ":3000"

# ==================== Snowflake ID Generator ====================
[snowflake]
machine_id = 1  # Machine ID (0-1023), must be unique in distributed environment

# ==================== Database Configuration (Required) ====================
# You can configure multiple databases, first one will be the default
[database.default]
host = "localhost"
port = 5432
user = "crab"
password = "crab"
db_name = "crab"
auto_migrate = true   # Auto migrate database schema
show_sql = false      # Show SQL logs

# Example: Additional database
# [database.usercenter]
# host = "localhost"
# port = 5432
# user = "crab"
# password = "crab"
# db_name = "crab_usercenter"
# auto_migrate = true
# show_sql = false

# ==================== Redis Configuration (Required) ====================
# Support multiple Redis instances, similar to database configuration
[redis.default]
addr = "localhost:6379"
password = ""
db = 0
# Cluster mode (automatically switches when configured)
# cluster = "host1:6379,host2:6379,host3:6379"

# Example: Additional Redis instance for caching
# [redis.cache]
# addr = "localhost:6380"
# password = ""
# db = 0

# ==================== Message Queue Configuration (Optional) ====================
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

# ==================== JWT Configuration (Optional) ====================
[jwt]
secret = "your-jwt-secret-change-me"
expire = "24h"

# ==================== Tracing Configuration (Optional) ====================
[trace]
service_name = "crab"
endpoint = ""  # Leave empty to disable, e.g.: localhost:4318
insecure = true

# ==================== Metrics Configuration (Optional) ====================
[metrics]
enabled = false
path = "/metrics"

# ==================== Storage Configuration (Optional) ====================
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
# Define different service combinations, start with: serve -s <name>

[[services]]
name = "all"
addr = ":3000"
modules = ["testapi", "ws"]

[[services]]
name = "api"
addr = ":3001"
modules = ["testapi"]

[[services]]
name = "ws"
addr = ":3002"
modules = ["ws"]
`
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
