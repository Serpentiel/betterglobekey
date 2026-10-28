package control

import (
	"cmp"
	"fmt"
	"time"

	domainconfig "github.com/Serpentiel/betterglobekey/internal/domain/config"
	controlv1 "github.com/Serpentiel/betterglobekey/internal/gen/betterglobekey/control/v1"
	"github.com/Serpentiel/betterglobekey/internal/infra/keymap"
)

// toProto converts the domain configuration model into its protobuf representation.
func toProto(cfg domainconfig.Config) *controlv1.Config {
	collections := make([]*controlv1.Collection, 0, len(cfg.Collections))
	for _, collection := range cfg.Collections {
		collections = append(collections, &controlv1.Collection{
			Name:    collection.Name,
			Sources: collection.Sources,
		})
	}

	return &controlv1.Config{
		Logger: &controlv1.Logger{
			Path:           cfg.Logger.Path,
			Level:          cfg.Logger.Level,
			RetentionDays:  int32(cfg.Logger.RetentionDays),  //nolint:gosec // retention values are small, non-negative
			RetentionFiles: int32(cfg.Logger.RetentionFiles), //nolint:gosec // retention values are small, non-negative
		},
		DoublePress: &controlv1.DoublePress{
			Enabled:      cfg.DoublePress.Enabled,
			MaximumDelay: cfg.DoublePress.MaxDelay.String(),
		},
		Reverse: &controlv1.Reverse{
			Enabled:  cfg.Reverse.Enabled,
			Modifier: cfg.Reverse.Modifier,
		},
		Hud: &controlv1.Hud{
			Enabled:        cfg.HUD.Enabled,
			Duration:       cfg.HUD.Duration.String(),
			ShowCollection: cfg.HUD.ShowCollection,
		},
		Collections: collections,
		Trigger:     triggerToProto(cfg.Trigger),
	}
}

// fromProto converts a protobuf configuration into the domain model, parsing the
// duration fields. Validation of the resulting model is left to the caller.
func fromProto(proto *controlv1.Config) (domainconfig.Config, error) {
	if proto == nil {
		return domainconfig.Config{}, fmt.Errorf("config is required")
	}

	delay, err := time.ParseDuration(proto.GetDoublePress().GetMaximumDelay())
	if err != nil {
		return domainconfig.Config{}, fmt.Errorf(
			"invalid double_press.maximum_delay %q: %w", proto.GetDoublePress().GetMaximumDelay(), err,
		)
	}

	duration, err := time.ParseDuration(proto.GetHud().GetDuration())
	if err != nil {
		return domainconfig.Config{}, fmt.Errorf("invalid hud.duration %q: %w", proto.GetHud().GetDuration(), err)
	}

	collections := make([]domainconfig.Collection, 0, len(proto.GetCollections()))
	for _, collection := range proto.GetCollections() {
		collections = append(collections, domainconfig.Collection{
			Name:    collection.GetName(),
			Sources: collection.GetSources(),
		})
	}

	return domainconfig.Config{
		Logger: domainconfig.Logger{
			Path:           proto.GetLogger().GetPath(),
			Level:          proto.GetLogger().GetLevel(),
			RetentionDays:  int(proto.GetLogger().GetRetentionDays()),
			RetentionFiles: int(proto.GetLogger().GetRetentionFiles()),
		},
		DoublePress: domainconfig.DoublePress{Enabled: proto.GetDoublePress().GetEnabled(), MaxDelay: delay},
		Reverse: domainconfig.Reverse{
			Enabled:  proto.GetReverse().GetEnabled(),
			Modifier: proto.GetReverse().GetModifier(),
		},
		HUD: domainconfig.HUD{
			Enabled:        proto.GetHud().GetEnabled(),
			Duration:       duration,
			ShowCollection: proto.GetHud().GetShowCollection(),
		},
		Trigger:     triggerFromProto(proto.GetTrigger()),
		Collections: collections,
	}, nil
}

// triggerToProto converts the trigger, every keyboard's included.
func triggerToProto(trigger domainconfig.Trigger) *controlv1.Trigger {
	keyboards := make(map[string]*controlv1.KeyboardTrigger, len(trigger.Keyboards))
	for id, keyboard := range trigger.Keyboards {
		keyboards[id] = keyboardTriggerToProto(keyboard)
	}

	return &controlv1.Trigger{Keyboards: keyboards}
}

func keyboardTriggerToProto(trigger domainconfig.KeyboardTrigger) *controlv1.KeyboardTrigger {
	return &controlv1.KeyboardTrigger{Key: trigger.Key, Mode: trigger.Mode}
}

// triggerFromProto converts a protobuf trigger, defaulting what a client that
// predates a field leaves empty: the Globe key and fn mode.
func triggerFromProto(proto *controlv1.Trigger) domainconfig.Trigger {
	var trigger domainconfig.Trigger

	if len(proto.GetKeyboards()) > 0 {
		trigger.Keyboards = make(map[string]domainconfig.KeyboardTrigger, len(proto.GetKeyboards()))
		for id, keyboard := range proto.GetKeyboards() {
			trigger.Keyboards[id] = keyboardTriggerFromProto(keyboard.GetKey(), keyboard.GetMode())
		}
	}

	return trigger
}

func keyboardTriggerFromProto(key, mode string) domainconfig.KeyboardTrigger {
	return domainconfig.KeyboardTrigger{
		Key:  cmp.Or(key, domainconfig.TriggerGlobe),
		Mode: cmp.Or(mode, domainconfig.TriggerModeFn),
	}
}

// keyboardsToProto converts the remapper's keyboard report.
func keyboardsToProto(statuses []keymap.Status, systemWide []keymap.Mapping) *controlv1.ListKeyboardsResponse {
	keyboards := make([]*controlv1.Keyboard, 0, len(statuses))

	for _, status := range statuses {
		keyboard := &controlv1.Keyboard{
			Id:            status.KeyboardID(),
			Name:          status.Name,
			Transport:     status.Transport,
			BuiltIn:       status.BuiltIn,
			HasGlobe:      status.HasGlobe(),
			Trigger:       keyboardTriggerToProto(status.Trigger),
			Listed:        status.Listed,
			OtherMappings: mappingsToProto(status.Others),
		}

		if status.Ours != nil {
			keyboard.TriggerMapping = mappingToProto(*status.Ours)
		}

		keyboards = append(keyboards, keyboard)
	}

	return &controlv1.ListKeyboardsResponse{Keyboards: keyboards, SystemMappings: mappingsToProto(systemWide)}
}

func mappingsToProto(mappings []keymap.Mapping) []*controlv1.KeyMapping {
	converted := make([]*controlv1.KeyMapping, 0, len(mappings))
	for _, mapping := range mappings {
		converted = append(converted, mappingToProto(mapping))
	}

	return converted
}

func mappingToProto(mapping keymap.Mapping) *controlv1.KeyMapping {
	return &controlv1.KeyMapping{
		Src:     mapping.Src,
		Dst:     mapping.Dst,
		SrcName: keymap.UsageName(mapping.Src),
		DstName: keymap.UsageName(mapping.Dst),
	}
}
