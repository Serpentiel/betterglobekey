import { contextBridge, ipcRenderer } from 'electron'

import type { Api, Config } from '../shared/types'

const api: Api = {
  getConfig: () => ipcRenderer.invoke('config:get'),
  applyConfig: (config: Config) => ipcRenderer.invoke('config:apply', config),
  listInputSources: () => ipcRenderer.invoke('sources:list'),
  getCurrentSource: () => ipcRenderer.invoke('sources:current'),
}

contextBridge.exposeInMainWorld('api', api)
