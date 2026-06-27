#import <Cocoa/Cocoa.h>

// A reusable borderless HUD panel that shows the active input source briefly,
// styled like the system bezel, with the collection name as a subtitle. All UI
// work happens on the main thread; show requests from other threads are
// marshalled via the main queue.

static NSPanel *hudWindow = nil;
static NSTextField *hudLabel = nil;
static NSTextField *hudSubLabel = nil;
static NSInteger hudShowToken = 0;

static const CGFloat hudWidth = 280.0;
static const CGFloat hudHeight = 104.0;
static const CGFloat hudCornerRadius = 18.0;
static double hudVisibleSeconds = 0.9;
static const double hudFadeSeconds = 0.35;

// hudSetDuration sets how long the HUD stays fully visible before fading.
void hudSetDuration(double seconds) {
	if (seconds > 0) {
		hudVisibleSeconds = seconds;
	}
}

static void hudEnsureWindow(void) {
	NSRect frame = NSMakeRect(0, 0, hudWidth, hudHeight);

	hudWindow = [[NSPanel alloc] initWithContentRect:frame
	                                       styleMask:NSWindowStyleMaskBorderless | NSWindowStyleMaskNonactivatingPanel
	                                         backing:NSBackingStoreBuffered
	                                           defer:NO];
	[hudWindow setLevel:NSStatusWindowLevel];
	[hudWindow setOpaque:NO];
	[hudWindow setBackgroundColor:[NSColor clearColor]];
	[hudWindow setIgnoresMouseEvents:YES];
	[hudWindow setHasShadow:YES];
	[hudWindow setCollectionBehavior:NSWindowCollectionBehaviorCanJoinAllSpaces |
	                                  NSWindowCollectionBehaviorStationary |
	                                  NSWindowCollectionBehaviorIgnoresCycle];

	NSVisualEffectView *blur = [[NSVisualEffectView alloc] initWithFrame:frame];
	blur.material = NSVisualEffectMaterialHUDWindow;
	blur.state = NSVisualEffectStateActive;
	blur.blendingMode = NSVisualEffectBlendingModeBehindWindow;
	blur.wantsLayer = YES;
	blur.layer.cornerRadius = hudCornerRadius;
	blur.layer.masksToBounds = YES;
	[hudWindow setContentView:blur];

	hudLabel = [[NSTextField alloc] initWithFrame:NSMakeRect(16, 48, hudWidth - 32, 34)];
	[hudLabel setEditable:NO];
	[hudLabel setSelectable:NO];
	[hudLabel setBordered:NO];
	[hudLabel setDrawsBackground:NO];
	[hudLabel setAlignment:NSTextAlignmentCenter];
	[hudLabel setTextColor:[NSColor whiteColor]];
	[hudLabel setFont:[NSFont systemFontOfSize:20 weight:NSFontWeightSemibold]];
	[[hudLabel cell] setLineBreakMode:NSLineBreakByTruncatingTail];
	[blur addSubview:hudLabel];

	hudSubLabel = [[NSTextField alloc] initWithFrame:NSMakeRect(16, 24, hudWidth - 32, 20)];
	[hudSubLabel setEditable:NO];
	[hudSubLabel setSelectable:NO];
	[hudSubLabel setBordered:NO];
	[hudSubLabel setDrawsBackground:NO];
	[hudSubLabel setAlignment:NSTextAlignmentCenter];
	[hudSubLabel setTextColor:[NSColor colorWithWhite:1.0 alpha:0.6]];
	[hudSubLabel setFont:[NSFont systemFontOfSize:13 weight:NSFontWeightRegular]];
	[[hudSubLabel cell] setLineBreakMode:NSLineBreakByTruncatingTail];
	[blur addSubview:hudSubLabel];
}

void hudRun(void) {
	@autoreleasepool {
		[NSApplication sharedApplication];
		[NSApp setActivationPolicy:NSApplicationActivationPolicyProhibited];
		[NSApp run];
	}
}

void hudStop(void) {
	dispatch_async(dispatch_get_main_queue(), ^{
		[NSApp stop:nil];
		// [NSApp run] only checks its stop flag after processing an event, so post one.
		NSEvent *wake = [NSEvent otherEventWithType:NSEventTypeApplicationDefined
		                                   location:NSZeroPoint
		                              modifierFlags:0
		                                  timestamp:0
		                               windowNumber:0
		                                    context:nil
		                                    subtype:0
		                                      data1:0
		                                      data2:0];
		[NSApp postEvent:wake atStart:YES];
	});
}

void hudShow(const char *ctext, const char *csubtitle) {
	NSString *text = [[NSString alloc] initWithUTF8String:ctext];
	NSString *subtitle = [[NSString alloc] initWithUTF8String:csubtitle];

	dispatch_async(dispatch_get_main_queue(), ^{
		if (hudWindow == nil) {
			hudEnsureWindow();
		}

		[hudLabel setStringValue:text];
		[hudSubLabel setStringValue:subtitle];
		[hudSubLabel setHidden:(subtitle.length == 0)];

		NSScreen *screen = [NSScreen mainScreen];
		if (screen != nil) {
			NSRect visible = [screen frame];
			NSPoint origin = NSMakePoint(NSMidX(visible) - hudWidth / 2.0,
			                             NSMidY(visible) - hudHeight / 2.0);
			[hudWindow setFrameOrigin:origin];
		}

		[hudWindow setAlphaValue:1.0];
		[hudWindow orderFrontRegardless];

		NSInteger token = ++hudShowToken;

		dispatch_after(dispatch_time(DISPATCH_TIME_NOW, (int64_t)(hudVisibleSeconds * NSEC_PER_SEC)),
		               dispatch_get_main_queue(), ^{
			if (token != hudShowToken) {
				return;
			}

			[NSAnimationContext runAnimationGroup:^(NSAnimationContext *context) {
				context.duration = hudFadeSeconds;
				[[hudWindow animator] setAlphaValue:0.0];
			} completionHandler:^{
				if (token == hudShowToken) {
					[hudWindow orderOut:nil];
				}
			}];
		});
	});
}
