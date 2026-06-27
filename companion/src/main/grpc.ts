import { existsSync } from 'node:fs'
import { homedir } from 'node:os'
import { join } from 'node:path'

import { credentials, loadPackageDefinition, type ServiceClientConstructor } from '@grpc/grpc-js'
import { loadSync } from '@grpc/proto-loader'
import { app } from 'electron'

import type { Config, InputSource } from '../shared/types'

// protoPath resolves the control.proto contract both in development (read from
// the synced copy next to the build output) and when packaged (from resources).
function protoPath(): string {
  return app.isPackaged
    ? join(process.resourcesPath, 'proto', 'control.proto')
    : join(__dirname, '..', '..', 'proto', 'control.proto')
}

// socketPath returns the daemon's control socket in the user's home directory.
function socketPath(): string {
  return join(homedir(), '.betterglobekey.sock')
}

interface ConfigServiceClient {
  getConfig: UnaryMethod<Record<string, never>, { config: Config }>
  applyConfig: UnaryMethod<{ config: Config }, Record<string, never>>
  listInputSources: UnaryMethod<Record<string, never>, { sources?: InputSource[] }>
  getCurrentSource: UnaryMethod<Record<string, never>, { source: InputSource }>
  close: () => void
}

type UnaryMethod<TReq, TRes> = (request: TReq, callback: (error: Error | null, response: TRes) => void) => void

let cached: ConfigServiceClient | undefined

function client(): ConfigServiceClient {
  if (cached) {
    return cached
  }

  const definition = loadSync(protoPath(), {
    keepCase: false,
    longs: String,
    enums: String,
    defaults: true,
    oneofs: true,
  })

  const pkg = loadPackageDefinition(definition) as unknown as {
    betterglobekey: { control: { v1: { ConfigService: ServiceClientConstructor } } }
  }

  const ServiceClient = pkg.betterglobekey.control.v1.ConfigService

  cached = new ServiceClient(`unix://${socketPath()}`, credentials.createInsecure()) as unknown as ConfigServiceClient

  return cached
}

function unary<TReq, TRes>(
  method: (req: TReq, cb: (e: Error | null, r: TRes) => void) => void,
  request: TReq,
): Promise<TRes> {
  return new Promise((resolve, reject) => {
    method(request, (error, response) => {
      if (error) {
        reject(error)

        return
      }

      resolve(response)
    })
  })
}

// daemonRunning reports whether the daemon's control socket exists.
export function daemonRunning(): boolean {
  return existsSync(socketPath())
}

export async function getConfig(): Promise<Config> {
  const service = client()
  const response = await unary(service.getConfig.bind(service), {})

  return response.config
}

export async function applyConfig(config: Config): Promise<void> {
  const service = client()
  await unary(service.applyConfig.bind(service), { config })
}

export async function listInputSources(): Promise<InputSource[]> {
  const service = client()
  const response = await unary(service.listInputSources.bind(service), {})

  return response.sources ?? []
}

export async function getCurrentSource(): Promise<InputSource> {
  const service = client()
  const response = await unary(service.getCurrentSource.bind(service), {})

  return response.source
}
