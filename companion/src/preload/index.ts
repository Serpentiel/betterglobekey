import { contextBridge, ipcRenderer } from 'electron'

import type { Api, Config } from '../shared/types'

const api: Api = {
  getConfig: () => ipcRenderer.invoke('config:get'),
  applyConfig: (config: Config) => ipcRenderer.invoke('config:apply', config),
  listInputSources: () => ipcRenderer.invoke('sources:list'),
  getVersion: () => ipcRenderer.invoke('version:get'),
}

contextBridge.exposeInMainWorld('api', api)
