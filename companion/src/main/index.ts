import { join } from 'node:path'

import { app, BrowserWindow, ipcMain, Menu, type MenuItemConstructorOptions, shell } from 'electron'

import type { Config } from '../shared/types'
import { applyConfig, getConfig, getVersion, listInputSources } from './grpc'

const APP_NAME = 'betterglobekey-companion'

function createWindow(): BrowserWindow {
  const window = new BrowserWindow({
    width: 820,
    height: 760,
    minWidth: 560,
    minHeight: 520,
    show: false,
    title: 'betterglobekey',
    backgroundColor: '#000000',
    titleBarStyle: 'hiddenInset',
    trafficLightPosition: { x: 19, y: 18 },
    webPreferences: {
      preload: join(__dirname, '../preload/index.js'),
      sandbox: false,
      spellcheck: false,
    },
  })

  window.on('ready-to-show', () => window.show())

  // Open external links in the browser and never navigate the window away from
  // the app itself.
  window.webContents.setWindowOpenHandler(({ url }) => {
    void shell.openExternal(url)

    return { action: 'deny' }
  })

  window.webContents.on('will-navigate', (event) => event.preventDefault())

  if (process.env.ELECTRON_RENDERER_URL) {
    void window.loadURL(process.env.ELECTRON_RENDERER_URL)
  } else {
    void window.loadFile(join(__dirname, '../renderer/index.html'))
  }

  return window
}

function buildMenu(): void {
  // Keep reload/devtools only in development; a shipped app should not expose
  // web-page affordances.
  const viewMenu: MenuItemConstructorOptions = app.isPackaged
    ? { label: 'View', submenu: [{ role: 'togglefullscreen' }] }
    : { role: 'viewMenu' }

  const menu = Menu.buildFromTemplate([{ role: 'appMenu' }, { role: 'editMenu' }, viewMenu, { role: 'windowMenu' }])

  Menu.setApplicationMenu(menu)
}

function registerHandlers(): void {
  ipcMain.handle('config:get', () => getConfig())
  ipcMain.handle('config:apply', (_event, config: Config) => applyConfig(config))
  ipcMain.handle('sources:list', () => listInputSources())
  ipcMain.handle('version:get', () => getVersion())
}

app.setName(APP_NAME)

if (!app.requestSingleInstanceLock()) {
  app.quit()
} else {
  let mainWindow: BrowserWindow | undefined

  app.on('second-instance', () => {
    if (mainWindow) {
      if (mainWindow.isMinimized()) {
        mainWindow.restore()
      }

      mainWindow.focus()
    }
  })

  void app.whenReady().then(() => {
    if (!app.isPackaged) {
      app.dock?.setIcon(join(__dirname, '../../build/icon.png'))
    }

    buildMenu()
    registerHandlers()
    mainWindow = createWindow()

    app.on('activate', () => {
      if (BrowserWindow.getAllWindows().length === 0) {
        mainWindow = createWindow()
      }
    })
  })

  app.on('window-all-closed', () => {
    if (process.platform !== 'darwin') {
      app.quit()
    }
  })
}
