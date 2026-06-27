# Usage

**betterglobekey** reworks the Globe key around two ideas: cycling **sources** within a collection, and switching
between **collections**. Holding **Shift** reverses the action.

On startup it begins from the input source that is currently active, selecting the collection that contains it.

## Single Press

A single press of the Globe key moves to the **next** input source in the current collection, wrapping back to the first
after the last.

## Shift + Single Press

Holding **Shift** and pressing once jumps to the **previously used** input source — repeating it toggles between the two
most recent sources, like the macOS "Select the previous input source" shortcut. (Hold Shift before you tap.)

## Double Press

Pressing the Globe key twice within the `double_press.maximum_delay` interval switches to the **next collection**,
restoring whichever source you last used there.

## Shift + Double Press

Holding **Shift** and double-pressing switches to the **previous collection**.

## On-screen HUD

On every change, a brief overlay appears in the center of the screen showing the new input source, with the collection
name as a subtitle. It is enabled by default and can be turned off with `hud: false` in the
[configuration](configuration.md).

## Example

Suppose you define a `default` collection with U.S. and Russian for everyday use, and a `coding` collection with U.S.
and a programmer layout. A single press flips between the languages in your current collection; a double press jumps over
to the other collection; Shift + single press snaps back to the language you were just using.

Experiment with different collections and the Shift modifier to find the setup that best suits your workflow.
