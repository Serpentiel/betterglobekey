#import <Cocoa/Cocoa.h>

// A reusable borderless HUD panel that briefly shows the active input source as
// a compact translucent capsule near the top of the screen (just under the menu
// bar / notch), like the system "device connected" overlay, with the collection
// name as a subtitle. All UI work happens on the main thread; show requests
// from other threads are marshalled via the main queue.

static NSPanel *hudWindow = nil;
static NSTextField *hudLabel = nil;
static NSTextField *hudSubLabel = nil;
static NSInteger hudShowToken = 0;

static const CGFloat hudWidth = 230.0;
static const CGFloat hudHeight = 58.0;
static const CGFloat hudCornerRadius = 18.0;
// hudTopGap is the distance below the top of the usable screen (clearing the
// menu bar / notch); hudSlideOffset is how far the HUD slides in from above.
static const CGFloat hudTopGap = 16.0;
static const CGFloat hudSlideOffset = 10.0;
// hudOpacity is the resting opacity — translucent, like the system overlays.
static const double hudOpacity = 0.82;
static double hudVisibleSeconds = 0.9;
static const double hudRevealSeconds = 0.28;
static const double hudFadeSeconds = 0.35;

// hudSetDuration sets how long the HUD stays fully visible before fading.
void hudSetDuration(double seconds) {
  if (seconds > 0) {
    hudVisibleSeconds = seconds;
  }
}

// hudMaskImage builds a resizable rounded-rectangle mask. Masking the visual
// effect view (rather than rounding its layer) clips the vibrancy material
// cleanly, leaving the corners fully transparent instead of opaque.
static NSImage *hudMaskImage(CGFloat radius) {
  CGFloat edge = radius * 2.0 + 1.0;

  NSImage *image =
      [NSImage imageWithSize:NSMakeSize(edge, edge)
                     flipped:NO
              drawingHandler:^BOOL(NSRect rect) {
                [[NSColor blackColor] setFill];
                [[NSBezierPath bezierPathWithRoundedRect:rect
                                                 xRadius:radius
                                                 yRadius:radius] fill];
                return YES;
              }];

  image.capInsets = NSEdgeInsetsMake(radius, radius, radius, radius);
  image.resizingMode = NSImageResizingModeStretch;

  return image;
}

static void hudEnsureWindow(void) {
  NSRect frame = NSMakeRect(0, 0, hudWidth, hudHeight);

  hudWindow =
      [[NSPanel alloc] initWithContentRect:frame
                                 styleMask:NSWindowStyleMaskBorderless |
                                           NSWindowStyleMaskNonactivatingPanel
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
  blur.maskImage = hudMaskImage(hudCornerRadius);
  [hudWindow setContentView:blur];

  hudLabel =
      [[NSTextField alloc] initWithFrame:NSMakeRect(14, 29, hudWidth - 28, 22)];
  [hudLabel setEditable:NO];
  [hudLabel setSelectable:NO];
  [hudLabel setBordered:NO];
  [hudLabel setDrawsBackground:NO];
  [hudLabel setAlignment:NSTextAlignmentCenter];
  [hudLabel setTextColor:[NSColor whiteColor]];
  [hudLabel setFont:[NSFont systemFontOfSize:14 weight:NSFontWeightSemibold]];
  [[hudLabel cell] setLineBreakMode:NSLineBreakByTruncatingTail];
  [blur addSubview:hudLabel];

  hudSubLabel =
      [[NSTextField alloc] initWithFrame:NSMakeRect(14, 11, hudWidth - 28, 15)];
  [hudSubLabel setEditable:NO];
  [hudSubLabel setSelectable:NO];
  [hudSubLabel setBordered:NO];
  [hudSubLabel setDrawsBackground:NO];
  [hudSubLabel setAlignment:NSTextAlignmentCenter];
  [hudSubLabel setTextColor:[NSColor colorWithWhite:1.0 alpha:0.6]];
  [hudSubLabel setFont:[NSFont systemFontOfSize:11 weight:NSFontWeightRegular]];
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
    // [NSApp run] only checks its stop flag after processing an event, so post
    // one.
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

    BOOL hasSubtitle = subtitle.length > 0;
    [hudLabel setStringValue:text];
    [hudSubLabel setStringValue:subtitle];
    [hudSubLabel setHidden:!hasSubtitle];
    // Centre the title vertically when it stands alone.
    [hudLabel setFrame:(hasSubtitle ? NSMakeRect(14, 29, hudWidth - 28, 22)
                                    : NSMakeRect(14, (hudHeight - 22) / 2.0,
                                                 hudWidth - 28, 22))];

    // Top-centre, just under the menu bar / notch.
    CGFloat finalX = 0.0;
    CGFloat finalY = 0.0;
    NSScreen *screen = [NSScreen mainScreen];
    if (screen != nil) {
      NSRect visible = [screen visibleFrame];
      finalX = NSMidX(visible) - hudWidth / 2.0;
      finalY = NSMaxY(visible) - hudHeight - hudTopGap;
    }

    // Slide down and fade in from above only when first appearing; if it is
    // already on screen (e.g. a rapid second switch), just stay put.
    if (![hudWindow isVisible]) {
      [hudWindow setFrameOrigin:NSMakePoint(finalX, finalY + hudSlideOffset)];
      [hudWindow setAlphaValue:0.0];
    }

    [hudWindow orderFrontRegardless];

    [NSAnimationContext runAnimationGroup:^(NSAnimationContext *context) {
      context.duration = hudRevealSeconds;
      [[hudWindow animator] setAlphaValue:hudOpacity];
      [[hudWindow animator] setFrameOrigin:NSMakePoint(finalX, finalY)];
    }];

    NSInteger token = ++hudShowToken;

    dispatch_after(dispatch_time(DISPATCH_TIME_NOW,
                                 (int64_t)(hudVisibleSeconds * NSEC_PER_SEC)),
                   dispatch_get_main_queue(), ^{
                     if (token != hudShowToken) {
                       return;
                     }

                     [NSAnimationContext
                         runAnimationGroup:^(NSAnimationContext *context) {
                           context.duration = hudFadeSeconds;
                           [[hudWindow animator] setAlphaValue:0.0];
                         }
                         completionHandler:^{
                           if (token == hudShowToken) {
                             [hudWindow orderOut:nil];
                           }
                         }];
                   });
  });
}
