#!/usr/bin/env node
'use strict'

const { execFileSync } = require('child_process')
const path = require('path')

const PLATFORM_MAP = { darwin: 'darwin', linux: 'linux', win32: 'windows' }
const ARCH_MAP = { x64: 'x64', arm64: 'arm64' }

const platform = PLATFORM_MAP[process.platform]
const arch = ARCH_MAP[process.arch]

if (!platform || !arch) {
  console.error(`mln: unsupported platform ${process.platform}/${process.arch}`)
  console.error('Please install manually: https://github.com/playsthisgame/melon/releases')
  process.exit(1)
}

const pkgName = `@playsthisgame/melon-${platform}-${arch}`
const binaryName = process.platform === 'win32' ? 'mln.exe' : 'mln'

let binary
try {
  const pkgDir = path.dirname(require.resolve(`${pkgName}/package.json`))
  binary = path.join(pkgDir, binaryName)
} catch (e) {
  console.error(`mln: could not find platform package ${pkgName}`)
  console.error('Try reinstalling: npm install -g @playsthisgame/melon')
  process.exit(1)
}

try {
  execFileSync(binary, process.argv.slice(2), { stdio: 'inherit' })
} catch (e) {
  process.exit(e.status || 1)
}
