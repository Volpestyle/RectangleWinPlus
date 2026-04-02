# RectangleWinPlus

A hotkey-oriented window snapping and resizing tool for Windows, with
**per-monitor virtual desktops**.

Forked from [ahmetb/RectangleWin](https://github.com/ahmetb/RectangleWin)
(a Windows rewrite of macOS [Rectangle.app](https://rectangleapp.com)) to add
per-monitor virtual desktops, multi-monitor support, and additional snap zones.

## Per-Monitor Virtual Desktops

RectangleWinPlus adds independent virtual desktops for each monitor — like macOS
Spaces, but on Windows. Unlike Windows' built-in virtual desktops (which switch
all monitors together), these let you switch desktops on one monitor without
affecting the others.

A HUD overlay shows the current desktop number when switching.

- **Switch desktop** on current monitor: <kbd>Ctrl</kbd> + <kbd>Win</kbd> + <kbd>&uarr;</kbd><kbd>&darr;</kbd>
- **Create new desktop** on current monitor: <kbd>Ctrl</kbd> + <kbd>Shift</kbd> + <kbd>Win</kbd> + <kbd>&uarr;</kbd>
- **Close current desktop** on current monitor: <kbd>Ctrl</kbd> + <kbd>Shift</kbd> + <kbd>Win</kbd> + <kbd>&darr;</kbd>
- **Move window to another monitor**: <kbd>Ctrl</kbd> + <kbd>Alt</kbd> + <kbd>&rarr;</kbd><kbd>&larr;</kbd>

Desktops are created and destroyed on demand — there's no fixed limit. Closing a
desktop merges its windows into the neighboring desktop. Move windows between
monitors with <kbd>Ctrl</kbd>+<kbd>Alt</kbd>+<kbd>&rarr;</kbd>/<kbd>&larr;</kbd>
— the window joins whichever desktop is active on the target monitor.

### Using with Windows Virtual Desktops

These per-monitor desktops operate as a subsystem inside Windows' own virtual
desktops. Windows desktops (<kbd>Win</kbd>+<kbd>Ctrl</kbd>+<kbd>&larr;</kbd><kbd>&rarr;</kbd>)
switch all monitors together; RectangleWinPlus desktops switch each monitor
independently. You can use both together for a full 3D grid of workspaces:
Windows desktops on one axis, per-monitor desktops on the other.

> **Safety:** If RectangleWinPlus crashes, hidden windows are automatically restored
> on next launch. You can also click **Show All Windows** in the system tray to
> manually restore everything.

## Window Snapping

This animation illustrates how RectangleWinPlus helps me move windows to edges
and corners (and cycle through half, one-thirds or two thirds width or height)
only using hotkeys:

![RectangleWinPlus demo](./assets/RectangleWin-demo.gif)

### Keyboard Bindings

- **Snap to edges** (left/right/top/bottom ½, ⅔, ⅓):
  - <kbd>Shift</kbd> + <kbd>Alt</kbd> + <kbd>&larr;</kbd><kbd>&rarr;</kbd><kbd>&uarr;</kbd><kbd>&darr;</kbd>
  - Press multiple times to alternate between ½, ⅔ and ⅓.

- **Corner snapping**
  - <kbd>Ctrl</kbd> + <kbd>Shift</kbd> + <kbd>Alt</kbd> + <kbd>&larr;</kbd>: top-left ½, ⅔ and ⅓
  - <kbd>Ctrl</kbd> + <kbd>Shift</kbd> + <kbd>Alt</kbd> + <kbd>&uarr;</kbd>: top-right ½, ⅔ and ⅓
  - <kbd>Ctrl</kbd> + <kbd>Shift</kbd> + <kbd>Alt</kbd> + <kbd>&darr;</kbd>: bottom-left ½, ⅔ and ⅓
  - <kbd>Ctrl</kbd> + <kbd>Shift</kbd> + <kbd>Alt</kbd> + <kbd>&rarr;</kbd>: bottom-right ½, ⅔ and ⅓

- **Sixth snapping** (3x2 grid):
  - <kbd>Shift</kbd> + <kbd>Alt</kbd> + <kbd>V</kbd>: top-left sixth
  - <kbd>Shift</kbd> + <kbd>Alt</kbd> + <kbd>B</kbd>: top-center sixth
  - <kbd>Shift</kbd> + <kbd>Alt</kbd> + <kbd>N</kbd>: top-right sixth
  - <kbd>Shift</kbd> + <kbd>Alt</kbd> + <kbd>M</kbd>: bottom-left sixth
  - <kbd>Shift</kbd> + <kbd>Alt</kbd> + <kbd>,</kbd>: bottom-center sixth
  - <kbd>Shift</kbd> + <kbd>Alt</kbd> + <kbd>.</kbd>: bottom-right sixth

- **Center third** (center ⅓ width, full height): <kbd>Shift</kbd>+<kbd>Alt</kbd>+<kbd>E</kbd>

- **Center window** on the display: <kbd>Win</kbd>+<kbd>Alt</kbd>+<kbd>C</kbd>

- **Maximize window**: <kbd>Win</kbd>+<kbd>Shift</kbd>+<kbd>F</kbd>

- **Always On Top (toggle)**: <kbd>Win</kbd>+<kbd>Alt</kbd>+<kbd>A</kbd>

- **Move to display** (next/previous monitor):
  - <kbd>Ctrl</kbd> + <kbd>Alt</kbd> + <kbd>&rarr;</kbd><kbd>&larr;</kbd>

## Install

1. Go to [Releases](https://github.com/Volpestyle/RectangleWinPlus/releases) and
   download the suitable binary for your architecture (typically x64).

2. Launch the `.exe` file. Now the program icon should be visible on system
   tray!

3. Click on the icon and mark as "Run on startup" to make sure you don't have
   to run it every time you reboot your PC.

## Why?

It seems that no window snapping utility for Windows is capable of letting
user snap windows to edges or corners in {half, two-thirds, one-third} sizes
using configurable **shortcut keys**, and center windows in a screen like
Rectangle.app does, so I wrote this small utility for myself.

I've tried the native Windows shortcuts and PowerToys FancyZones and they
are not supporting corners, alternating between half and one/two thirds, and
are not offering enough hotkey support.

## Roadmap

- Configurable shortcuts: I don't need these and it will likely require a pop-up
  UI, so I will probably not get to this.

## Development (Install from source)

With Go 1.17+ installed, clone this repository and run:

```sh
go generate
GOOS=windows go build -ldflags -H=windowsgui .
```

The `RectangleWinPlus.exe` will be available in the same directory.

## License

This project is distributed as-is under the Apache 2.0 license.
See [LICENSE](./LICENSE).

If you see bugs, please open issues. I can't promise any fixes.
