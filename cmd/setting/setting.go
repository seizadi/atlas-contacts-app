package setting

import (
	"path"
	"os"
	"fmt"
	
	"gopkg.in/ini.v1"
	"strings"
	"path/filepath"
	"net/url"
	"regexp"
)

type CommandLineArgs struct {
	Config   string
	HomePath string
	Args     []string
}

var (
	// Paths
	HomePath         string
	CustomInitPath   = "conf/custom.ini"
	
	// Global setting objects.
	Cfg          *ini.File
	
	// Global for logging
	configFiles                  []string
	appliedCommandLineProperties []string
	appliedEnvOverrides          []string
)

func NewConfigContext(args *CommandLineArgs) error {
	setHomePath(args)
	loadConfiguration(args)
	return nil
}


func setHomePath(args *CommandLineArgs) {
	if args.HomePath != "" {
		HomePath = args.HomePath
		return
	}
	
	HomePath, _ = filepath.Abs(".")
	// check if homepath is correct
	if pathExists(filepath.Join(HomePath, "conf/defaults.ini")) {
		return
	}
	
	// try down one path
	if pathExists(filepath.Join(HomePath, "../conf/defaults.ini")) {
		HomePath = filepath.Join(HomePath, "../")
	}
	
	fmt.Printf("Atlas Server Init: Failed to set Home Path %s.", HomePath)
	os.Exit(1)
}

func loadConfiguration(args *CommandLineArgs) {
	var err error
	
	// load config defaults
	defaultConfigFile := path.Join(HomePath, "conf/defaults.ini")
	configFiles = append(configFiles, defaultConfigFile)
	
	// check if config file exists
	if _, err := os.Stat(defaultConfigFile); os.IsNotExist(err) {
		fmt.Printf("Atlas Server Init: Failed could not find config defaults using %s\n", defaultConfigFile)
		os.Exit(1)
	}
	
	// load defaults
	Cfg, err = ini.Load(defaultConfigFile)
	if err != nil {
		fmt.Println(fmt.Sprintf("Atlas Server Init: Failed to parse defaults.ini, %v", err))
		os.Exit(1)
		return
	}
	
	Cfg.BlockMode = false
	
	// command line props
	commandLineProps := getCommandLineProperties(args.Args)
	// load default overrides
	applyCommandLineDefaultProperties(commandLineProps)
	
	// load specified config file
	err = loadSpecifedConfigFile(args.Config)
	if err != nil {
		fmt.Println(fmt.Sprintf("Atlas Server Init: Failed to load config %s, %v", args.Config, err))
		os.Exit(1)
		return
	}
	
	// apply environment overrides
	applyEnvVariableOverrides()
	
	// apply command line overrides
	applyCommandLineProperties(commandLineProps)
	
	// evaluate config values containing environment variables
	evalConfigValues()
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func getCommandLineProperties(args []string) map[string]string {
	props := make(map[string]string)
	
	for _, arg := range args {
		if !strings.HasPrefix(arg, "cfg:") {
			continue
		}
		
		trimmed := strings.TrimPrefix(arg, "cfg:")
		parts := strings.Split(trimmed, "=")
		if len(parts) != 2 {
			fmt.Printf("Atlas Server Init: Invalid command line argument %s", arg)
			os.Exit(1)
			return nil
		}
		
		props[parts[0]] = parts[1]
	}
	return props
}

func applyCommandLineDefaultProperties(props map[string]string) {
	appliedCommandLineProperties = make([]string, 0)
	for _, section := range Cfg.Sections() {
		for _, key := range section.Keys() {
			keyString := fmt.Sprintf("default.%s.%s", section.Name(), key.Name())
			value, exists := props[keyString]
			if exists {
				key.SetValue(value)
				if shouldRedactKey(keyString) {
					value = "*********"
				}
				appliedCommandLineProperties = append(appliedCommandLineProperties, fmt.Sprintf("%s=%s", keyString, value))
			}
		}
	}
}

func shouldRedactKey(s string) bool {
	uppercased := strings.ToUpper(s)
	return strings.Contains(uppercased, "PASSWORD") || strings.Contains(uppercased, "SECRET")
}

func shouldRedactURLKey(s string) bool {
	uppercased := strings.ToUpper(s)
	return strings.Contains(uppercased, "SECRET_URL")
}

func loadSpecifedConfigFile(configFile string) error {
	if configFile == "" {
		configFile = filepath.Join(HomePath, CustomInitPath)
		// return without error if custom file does not exist
		if !pathExists(configFile) {
			return nil
		}
	}
	
	userConfig, err := ini.Load(configFile)
	if err != nil {
		return fmt.Errorf("Failed to parse %v, %v", configFile, err)
	}
	
	userConfig.BlockMode = false
	
	for _, section := range userConfig.Sections() {
		for _, key := range section.Keys() {
			if key.Value() == "" {
				continue
			}
			
			defaultSec, err := Cfg.GetSection(section.Name())
			if err != nil {
				defaultSec, _ = Cfg.NewSection(section.Name())
			}
			defaultKey, err := defaultSec.GetKey(key.Name())
			if err != nil {
				defaultKey, _ = defaultSec.NewKey(key.Name(), key.Value())
			}
			defaultKey.SetValue(key.Value())
		}
	}
	
	configFiles = append(configFiles, configFile)
	return nil
}

func applyEnvVariableOverrides() {
	appliedEnvOverrides = make([]string, 0)
	for _, section := range Cfg.Sections() {
		for _, key := range section.Keys() {
			sectionName := strings.ToUpper(strings.Replace(section.Name(), ".", "_", -1))
			keyName := strings.ToUpper(strings.Replace(key.Name(), ".", "_", -1))
			envKey := fmt.Sprintf("ATLAS_%s_%s", sectionName, keyName)
			envValue := os.Getenv(envKey)
			
			if len(envValue) > 0 {
				key.SetValue(envValue)
				if shouldRedactKey(envKey) {
					envValue = "*********"
				}
				if shouldRedactURLKey(envKey) {
					u, _ := url.Parse(envValue)
					ui := u.User
					if ui != nil {
						_, exists := ui.Password()
						if exists {
							u.User = url.UserPassword(ui.Username(), "-redacted-")
							envValue = u.String()
						}
					}
				}
				appliedEnvOverrides = append(appliedEnvOverrides, fmt.Sprintf("%s=%s", envKey, envValue))
			}
		}
	}
}

func applyCommandLineProperties(props map[string]string) {
	for _, section := range Cfg.Sections() {
		sectionName := section.Name() + "."
		if section.Name() == ini.DEFAULT_SECTION {
			sectionName = ""
		}
		for _, key := range section.Keys() {
			keyString := sectionName + key.Name()
			value, exists := props[keyString]
			if exists {
				appliedCommandLineProperties = append(appliedCommandLineProperties, fmt.Sprintf("%s=%s", keyString, value))
				key.SetValue(value)
			}
		}
	}
}

func evalConfigValues() {
	for _, section := range Cfg.Sections() {
		for _, key := range section.Keys() {
			key.SetValue(evalEnvVarExpression(key.Value()))
		}
	}
}

func evalEnvVarExpression(value string) string {
	regex := regexp.MustCompile(`\${(\w+)}`)
	return regex.ReplaceAllStringFunc(value, func(envVar string) string {
		envVar = strings.TrimPrefix(envVar, "${")
		envVar = strings.TrimSuffix(envVar, "}")
		envValue := os.Getenv(envVar)
		
		// if env variable is hostname and it is empty use os.Hostname as default
		if envVar == "HOSTNAME" && envValue == "" {
			envValue, _ = os.Hostname()
		}
		
		return envValue
	})
}