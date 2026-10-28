import { getVersion as getCompanionVersion } from '@tauri-apps/api/app'
import { invoke } from '@tauri-apps/api/core'

import type { Api, Config } from '../../../shared/types'

/** api calls the daemon through the Tauri commands in src-tauri/src/lib.rs, and
 * reads the companion's own version from Tauri. */
export const api: Api = {
  getConfig: () => invoke('get_config'),
  applyConfig: (config: Config) => invoke('apply_config', { config }),
  listInputSources: () => invoke('list_input_sources'),
  listKeyboards: () => invoke('list_keyboards'),
  recordKey: (keyboardId: string) => invoke('record_key', { keyboardId }),
  openInputMonitoringSettings: () => invoke('open_input_monitoring_settings'),
  getVersion: () => invoke('get_version'),
  getCompanionVersion,
}
