package keymap

import "fmt"

// usageNames names the HID usages mappings commonly involve.
var usageNames = map[uint64]string{
	GlobeUsage:          "fn",
	usageCapsLock:       "Caps Lock",
	usageApplication:    "Menu",
	usageLeftControl:    "Left Control",
	usageLeftShift:      "Left Shift",
	usageLeftOption:     "Left Option",
	usageLeftCommand:    "Left Command",
	usageRightControl:   "Right Control",
	usageRightShift:     "Right Shift",
	usageRightOption:    "Right Option",
	usageRightCommand:   "Right Command",
	keyboardPage | 0x28: "Return",
	keyboardPage | 0x29: "Escape",
	keyboardPage | 0x2A: "Delete",
	keyboardPage | 0x2B: "Tab",
	keyboardPage | 0x2C: "Space",
	keyboardPage | 0x35: "Grave Accent",
	keyboardPage | 0x46: "Print Screen",
	keyboardPage | 0x47: "Scroll Lock",
	keyboardPage | 0x48: "Pause",
	keyboardPage | 0x49: "Insert",
	keyboardPage | 0x4A: "Home",
	keyboardPage | 0x4B: "Page Up",
	keyboardPage | 0x4C: "Forward Delete",
	keyboardPage | 0x4D: "End",
	keyboardPage | 0x4E: "Page Down",
	keyboardPage | 0x4F: "Right Arrow",
	keyboardPage | 0x50: "Left Arrow",
	keyboardPage | 0x51: "Down Arrow",
	keyboardPage | 0x52: "Up Arrow",
	keyboardPage | 0x64: "Section",
}

// usageMask selects the usage from a usage page and usage.
const usageMask = 0xFFFFFFFF

// firstExtendedF is the number of the first function key after F12's range.
const firstExtendedF = 13

// Keyboard page ranges named by pattern.
const (
	keyA   = 0x04
	keyZ   = 0x1D
	key1   = 0x1E
	key0   = 0x27
	keyF1  = 0x3A
	keyF12 = 0x45
	keyF13 = 0x68
	keyF24 = 0x73
)

// letters and digits are the keys of the A-Z and 1-0 usage ranges, in order.
const (
	letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits  = "1234567890"
)

// KeyUsage returns a usage without its usage page.
func KeyUsage(usage uint64) uint32 {
	return uint32(usage & usageMask) //nolint:gosec // masked to 32 bits
}

// UsageName names a HID usage, or formats it in hex.
func UsageName(usage uint64) string {
	if name, ok := usageNames[usage]; ok {
		return name
	}

	if usage>>32 == keyboardPage>>32 {
		switch key := usage & usageMask; {
		case key >= keyA && key <= keyZ:
			return letters[key-keyA : key-keyA+1]
		case key >= key1 && key <= key0:
			return digits[key-key1 : key-key1+1]
		case key >= keyF1 && key <= keyF12:
			return fmt.Sprintf("F%d", key-keyF1+1)
		case key >= keyF13 && key <= keyF24:
			return fmt.Sprintf("F%d", key-keyF13+firstExtendedF)
		}
	}

	return fmt.Sprintf("%#x", usage)
}
