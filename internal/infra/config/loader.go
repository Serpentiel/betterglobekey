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

	enabled := true

	schema.Version = currentVersion
	schema.Logger.Path = defaultLogPath()
	schema.Logger.Retention.Days = defaultRetentionDays
	schema.Logger.Retention.Files = defaultRetentionFiles
	schema.DoublePress.MaximumDelay = defaultDoublePressDelay
	schema.HUD = &enabled

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

// toDomain converts a parsed v2 schema into the domain configuration model.
func toDomain(schema schemaV2) (config.Config, error) {
	delay, err := time.ParseDuration(schema.DoublePress.MaximumDelay)
	if err != nil {
		return config.Config{}, fmt.Errorf(
			"invalid double_press.maximum_delay %q: %w", schema.DoublePress.MaximumDelay, err,
		)
	}

	cfg := config.Config{
		Logger: config.Logger{
			Path:           schema.Logger.Path,
			RetentionDays:  schema.Logger.Retention.Days,
			RetentionFiles: schema.Logger.Retention.Files,
		},
		DoublePressMaxDelay: delay,
		HUD:                 schema.HUD == nil || *schema.HUD,
		Collections:         make([]config.Collection, 0, len(schema.Collections)),
	}

	for _, collection := range schema.Collections {
		cfg.Collections = append(cfg.Collections, config.Collection{Name: collection.Name, Sources: collection.Sources})
	}

	return cfg, nil
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
