package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Serpentiel/betterglobekey/internal/domain/config"
	"gopkg.in/yaml.v3"
)

// configFileMode is the permission mode used when writing the configuration file.
const configFileMode os.FileMode = 0o600

// defaultHeader is prepended to freshly generated configuration files.
const defaultHeader = "# betterglobekey configuration.\n" +
	"# Documentation: https://github.com/Serpentiel/betterglobekey/tree/main/docs\n\n"

// Load reads, migrates if necessary, and validates the configuration at path.
// When the on-disk schema is outdated it is upgraded in place, with the original
// file preserved as a versioned backup alongside it.
func Load(path string) (config.Config, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path is the user-provided config file
	if err != nil {
		return config.Config{}, fmt.Errorf("reading config %q: %w", path, err)
	}

	migrated, from, err := migrate(data)
	if err != nil {
		return config.Config{}, err
	}

	if from < currentVersion {
		if err = writeWithBackup(path, data, migrated, from); err != nil {
			return config.Config{}, err
		}
	}

	var schema schemaV2
	if err = yaml.Unmarshal(migrated, &schema); err != nil {
		return config.Config{}, fmt.Errorf("parsing config %q: %w", path, err)
	}

	cfg, err := toDomain(schema)
	if err != nil {
		return config.Config{}, err
	}

	if err = cfg.Validate(); err != nil {
		return config.Config{}, fmt.Errorf("invalid config %q: %w", path, err)
	}

	return cfg, nil
}

// WriteDefault generates a fresh v2 configuration at path, seeding the primary
// collection with the provided input sources.
func WriteDefault(path string, sources []string) error {
	var schema schemaV2

	schema.Version = currentVersion
	schema.Logger.Path = defaultLogPath()
	schema.Logger.Level = defaultLogLevel
	schema.Logger.Retention.Days = defaultRetentionDays
	schema.Logger.Retention.Files = defaultRetentionFiles
	schema.DoublePress.Enabled = boolPtr(true)
	schema.DoublePress.MaximumDelay = defaultDoublePressDelay
	schema.Reverse.Enabled = boolPtr(true)
	schema.Reverse.Modifier = defaultReverseModifier
	schema.HUD.Enabled = boolPtr(true)
	schema.HUD.Duration = defaultHUDDuration
	schema.HUD.ShowCollection = boolPtr(true)

	if len(sources) > 0 {
		schema.Collections = []collectionV2{{Name: defaultCollectionName, Sources: sources}}
	}

	data, err := yaml.Marshal(schema)
	if err != nil {
		return fmt.Errorf("encoding default config: %w", err)
	}

	//nolint:gosec // path is the user-provided config file
	if err := os.WriteFile(path, append([]byte(defaultHeader), data...), configFileMode); err != nil {
		return fmt.Errorf("writing default config %q: %w", path, err)
	}

	return nil
}

// writeWithBackup persists the migrated document, first preserving the original
// as a versioned backup.
func writeWithBackup(path string, original, migrated []byte, from int) error {
	backup := fmt.Sprintf("%s.v%d.bak", path, from)

	//nolint:gosec // path is the user-provided config file
	if err := os.WriteFile(backup, original, configFileMode); err != nil {
		return fmt.Errorf("backing up config to %q: %w", backup, err)
	}

	//nolint:gosec // path is the user-provided config file
	if err := os.WriteFile(path, migrated, configFileMode); err != nil {
		return fmt.Errorf("writing migrated config %q: %w", path, err)
	}

	return nil
}

// toDomain converts a parsed v2 schema into the domain configuration model,
// applying defaults for absent optional values.
func toDomain(schema schemaV2) (config.Config, error) {
	applySchemaDefaults(&schema)

	delay, err := time.ParseDuration(schema.DoublePress.MaximumDelay)
	if err != nil {
		return config.Config{}, fmt.Errorf(
			"invalid double_press.maximum_delay %q: %w", schema.DoublePress.MaximumDelay, err,
		)
	}

	duration, err := time.ParseDuration(schema.HUD.Duration)
	if err != nil {
		return config.Config{}, fmt.Errorf("invalid hud.duration %q: %w", schema.HUD.Duration, err)
	}

	cfg := config.Config{
		Logger: config.Logger{
			Path:           schema.Logger.Path,
			Level:          schema.Logger.Level,
			RetentionDays:  schema.Logger.Retention.Days,
			RetentionFiles: schema.Logger.Retention.Files,
		},
		DoublePress: config.DoublePress{Enabled: boolOr(schema.DoublePress.Enabled, true), MaxDelay: delay},
		Reverse:     config.Reverse{Enabled: boolOr(schema.Reverse.Enabled, true), Modifier: schema.Reverse.Modifier},
		HUD: config.HUD{
			Enabled:        boolOr(schema.HUD.Enabled, true),
			Duration:       duration,
			ShowCollection: boolOr(schema.HUD.ShowCollection, true),
		},
		Collections: make([]config.Collection, 0, len(schema.Collections)),
	}

	for _, collection := range schema.Collections {
		cfg.Collections = append(cfg.Collections, config.Collection{Name: collection.Name, Sources: collection.Sources})
	}

	return cfg, nil
}

// applySchemaDefaults fills empty string fields with their defaults so a
// partially-specified file still loads.
func applySchemaDefaults(schema *schemaV2) {
	if schema.Logger.Level == "" {
		schema.Logger.Level = defaultLogLevel
	}

	if schema.DoublePress.MaximumDelay == "" {
		schema.DoublePress.MaximumDelay = defaultDoublePressDelay
	}

	if schema.Reverse.Modifier == "" {
		schema.Reverse.Modifier = defaultReverseModifier
	}

	if schema.HUD.Duration == "" {
		schema.HUD.Duration = defaultHUDDuration
	}
}

// defaultLogPath returns the conventional macOS log location, falling back to a
// relative path if the home directory cannot be determined.
func defaultLogPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "betterglobekey.log"
	}

	return filepath.Join(home, "Library", "Logs", "betterglobekey.log")
}
