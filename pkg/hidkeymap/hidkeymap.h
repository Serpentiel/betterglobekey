#ifndef BGK_HIDKEYMAP_H
#define BGK_HIDKEYMAP_H

#include <stdint.h>

// bgkKeyboard describes a connected keyboard service.
typedef struct {
	// id is the service's registry entry ID.
	uint64_t id;
	uint32_t vendor;
	uint32_t product;
	uint32_t location;
	int builtIn;
	char name[128];
	char transport[32];
} bgkKeyboard;

// Results of the per-keyboard calls besides a count or success.
#define bgkUnavailable (-1)
#define bgkNotFound (-2)

int bgkKeymapGet(uint64_t *src, uint64_t *dst, int capacity);
int bgkKeymapSet(const uint64_t *src, const uint64_t *dst, int count);
int bgkKeyboardsList(bgkKeyboard *out, int capacity);
int bgkKeyboardKeymapGet(uint64_t id, uint64_t *src, uint64_t *dst, int capacity);
int bgkKeyboardKeymapSet(uint64_t id, const uint64_t *src, const uint64_t *dst, int count);

#endif
