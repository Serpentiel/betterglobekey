import type { Api } from '../shared/types'

declare global {
  interface Window {
    api: Api
  }
}

export {}
