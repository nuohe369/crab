package boot

import (
	"fmt"
	"log"
	"sort"

	"github.com/nuohe369/crab/common/config"
	"github.com/nuohe369/crab/pkg/pgsql"
)

// DBNamer interface allows models to specify their database
// DBNamer 接口允许模型指定其数据库
type DBNamer interface {
	DBName() string
}

// ModuleDependency represents a module's database dependencies
// ModuleDependency 表示模块的数据库依赖
type ModuleDependency struct {
	ModuleName string   // Module name | 模块名称
	Models     []any    // Models used by module | 模块使用的模型
	Databases  []string // Required databases (deduplicated) | 需要的数据库（去重后）
}

// CheckModuleDependencies analyzes a module's database dependencies
// CheckModuleDependencies 分析模块的数据库依赖
func CheckModuleDependencies(module Module) *ModuleDependency {
	models := module.Models()
	dbMap := make(map[string]bool)

	// Collect and deduplicate database dependencies
	// 收集并去重数据库依赖
	for _, model := range models {
		if namer, ok := model.(DBNamer); ok {
			dbName := namer.DBName()
			if dbName != "" {
				dbMap[dbName] = true
			} else {
				// Model has DBName() but returns empty, uses default database
				// 模型有 DBName() 但返回空，使用默认数据库
				dbMap["default"] = true
			}
		} else {
			// Model doesn't have DBName(), uses default database
			// 模型没有 DBName()，使用默认数据库
			dbMap["default"] = true
		}
	}

	// Convert map to sorted slice
	// 转换为排序的切片
	databases := make([]string, 0, len(dbMap))
	for db := range dbMap {
		databases = append(databases, db)
	}
	sort.Strings(databases)

	return &ModuleDependency{
		ModuleName: module.Name(),
		Models:     models,
		Databases:  databases,
	}
}

// ValidateModuleDependencies checks if all required databases are configured
// ValidateModuleDependencies 检查所有需要的数据库是否已配置
func ValidateModuleDependencies(dep *ModuleDependency) error {
	var missing []string
	var configured []string

	for _, dbName := range dep.Databases {
		var db *pgsql.Client
		if dbName == "default" {
			db = pgsql.Get()
		} else {
			db = pgsql.Get(dbName)
		}

		if db == nil {
			missing = append(missing, dbName)
		} else {
			configured = append(configured, dbName)
		}
	}

	if len(missing) > 0 {
		return &DependencyError{
			ModuleName:  dep.ModuleName,
			Missing:     missing,
			Configured:  configured,
			AllRequired: dep.Databases,
		}
	}

	return nil
}

// DependencyError represents a module dependency validation error
// DependencyError 表示模块依赖验证错误
type DependencyError struct {
	ModuleName  string
	Missing     []string
	Configured  []string
	AllRequired []string
}

// Error implements the error interface
// Error 实现 error 接口
func (e *DependencyError) Error() string {
	return fmt.Sprintf("module '%s' requires unconfigured databases: %v", e.ModuleName, e.Missing)
}

// GetConfiguredDatabases returns a map of configured database names
// GetConfiguredDatabases 返回已配置的数据库名称映射
func GetConfiguredDatabases() map[string]bool {
	databases := config.GetDatabases()
	configured := make(map[string]bool)

	// Check for default database
	// 检查默认数据库
	if pgsql.Get() != nil {
		configured["default"] = true
	}

	// Check for named databases
	// 检查命名数据库
	for name := range databases {
		if pgsql.Get(name) != nil {
			configured[name] = true
		}
	}

	return configured
}

// ValidateAndFilterModules validates module dependencies and returns only valid modules
// ValidateAndFilterModules 验证模块依赖并返回有效的模块
func ValidateAndFilterModules(modules []Module, strict bool) []Module {
	if len(modules) == 0 {
		return modules
	}

	log.Println("🔍 Checking module dependencies...")

	var validModules []Module
	var failedModules []string

	for _, module := range modules {
		// Check dependencies
		// 检查依赖
		dep := CheckModuleDependencies(module)

		// Log module dependencies
		// 记录模块依赖
		if len(dep.Databases) > 0 {
			log.Printf("  📦 Module '%s' requires databases: %v", module.Name(), dep.Databases)
		} else {
			log.Printf("  📦 Module '%s' has no database dependencies", module.Name())
		}

		// Validate dependencies
		// 验证依赖
		if err := ValidateModuleDependencies(dep); err != nil {
			depErr := err.(*DependencyError)

			log.Printf("  ❌ Module '%s' dependency check failed:", module.Name())
			if len(depErr.Configured) > 0 {
				log.Printf("     ✓ Configured: %v", depErr.Configured)
			}
			log.Printf("     ✗ Missing: %v", depErr.Missing)

			// Print configuration hints
			// 打印配置提示
			log.Printf("  💡 To enable this module, add to config.toml:")
			for _, dbName := range depErr.Missing {
				if dbName == "default" {
					log.Printf("     [database.default]")
					log.Printf("     host = \"localhost\"")
					log.Printf("     port = 5432")
					log.Printf("     user = \"your_user\"")
					log.Printf("     password = \"your_password\"")
					log.Printf("     db_name = \"your_database\"")
				} else {
					log.Printf("     [database.%s]", dbName)
					log.Printf("     host = \"localhost\"")
					log.Printf("     port = 5432")
					log.Printf("     user = \"%s\"", dbName)
					log.Printf("     password = \"%s\"", dbName)
					log.Printf("     db_name = \"%s\"", dbName)
				}
			}

			if strict {
				log.Printf("  ⚠️  Module '%s' will NOT be started (strict mode)", module.Name())
				failedModules = append(failedModules, module.Name())
				continue
			} else {
				log.Printf("  ⚠️  Module '%s' will be started anyway (non-strict mode)", module.Name())
				log.Printf("  ⚠️  Runtime errors may occur when accessing missing databases")
			}
		} else {
			log.Printf("  ✅ Module '%s' dependencies satisfied", module.Name())
		}

		validModules = append(validModules, module)
	}

	// Summary
	// 总结
	if len(failedModules) > 0 {
		log.Printf("⚠️  %d module(s) skipped due to missing dependencies: %v", len(failedModules), failedModules)
	}
	if len(validModules) > 0 {
		log.Printf("✅ %d module(s) ready to start", len(validModules))
	} else {
		log.Println("❌ No modules available to start")
	}

	return validModules
}

// PrintModuleDependencies prints a summary of all module dependencies
// PrintModuleDependencies 打印所有模块依赖的摘要
func PrintModuleDependencies(modules []Module) {
	log.Println("📊 Module Dependency Summary:")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	configured := GetConfiguredDatabases()

	for _, module := range modules {
		dep := CheckModuleDependencies(module)
		log.Printf("Module: %s", dep.ModuleName)

		if len(dep.Databases) == 0 {
			log.Println("  No database dependencies")
		} else {
			log.Println("  Required databases:")
			for _, dbName := range dep.Databases {
				status := "✗"
				if configured[dbName] {
					status = "✓"
				}
				log.Printf("    %s %s", status, dbName)
			}
		}

		if len(dep.Models) > 0 {
			log.Printf("  Models: %d", len(dep.Models))
		}
		log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	}
}
