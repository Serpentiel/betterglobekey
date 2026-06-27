import { join } from 'node:path'

import { app, BrowserWindow, ipcMain } from 'electron'

import type { Config } from '../shared/types'
import { applyConfig, getConfig, getCurrentSource, listInputSources } from './grpc'

function createWindow(): void {
  const window = new BrowserWindow({
    width: 820,
    height: 760,
    minWidth: 640,
    minHeight: 560,
    show: false,
    title: 'betterglobekey companion',
    titleBarStyle: 'hiddenInset',
    webPreferences: {
      preload: join(__dirname, '../preload/index.js'),
      sandbox: false,
    },
  })

  window.on('ready-to-show', () => window.show())

  if (process.env.ELECTRON_RENDERER_URL) {
    void window.loadURL(process.env.ELECTRON_RENDERER_URL)
  } else {
    void window.loadFile(join(__dirname, '../renderer/index.html'))
  }
}

function registerHandlers(): void {
  ipcMain.handle('config:get', () => getConfig())
  ipcMain.handle('config:apply', (_event, config: Config) => applyConfig(config))
  ipcMain.handle('sources:list', () => listInputSources())
  ipcMain.handle('sources:current', () => getCurrentSource())
}

void app.whenReady().then(() => {
  registerHandlers()
  createWindow()

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow()
    }
  })
})

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit()
  }
})
