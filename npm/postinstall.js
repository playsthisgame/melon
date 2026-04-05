#!/usr/bin/env node
'use strict'

const https = require('https')
const fs = require('fs')
const path = require('path')
const os = require('os')
const { execSync } = require('child_process')

const pkg = require('./package.json')
const version = pkg.version

const OS_MAP = { darwin: 'darwin', linux: 'linux', win32: 'windows' }
const ARCH_MAP = { x64: 'amd64', arm64: 'arm64' }

const goos = OS_MAP[process.platform]
const goarch = ARCH_MAP[process.arch]

if (!goos || !goarch) {
  console.error(`mln: unsupported platform ${process.platform}/${process.arch}`)
  console.error('Please install manually: https://github.com/playsthisgame/melon/releases')
  process.exit(1)
}

const ext = process.platform === 'win32' ? 'zip' : 'tar.gz'
const archiveName = `mln_${version}_${goos}_${goarch}.${ext}`
const downloadURL = `https://github.com/playsthisgame/melon/releases/download/v${version}/${archiveName}`
const archivePath = path.join(os.tmpdir(), archiveName)
const binaryName = process.platform === 'win32' ? 'mln.exe' : 'mln'
const binaryDest = path.join(__dirname, binaryName)

function download(url, dest, cb) {
  const file = fs.createWriteStream(dest)
  https.get(url, res => {
    if (res.statusCode === 301 || res.statusCode === 302) {
      file.close()
      return download(res.headers.location, dest, cb)
    }
    if (res.statusCode !== 200) {
      file.close()
      return cb(new Error(`HTTP ${res.statusCode} downloading ${url}`))
    }
    res.pipe(file)
    file.on('finish', () => file.close(cb))
  }).on('error', err => {
    file.close()
    fs.unlink(dest, () => {})
    cb(err)
  })
}

console.log(`mln: downloading ${archiveName}...`)

download(downloadURL, archivePath, err => {
  if (err) {
    console.error(`mln: download failed: ${err.message}`)
    process.exit(1)
  }

  try {
    if (process.platform === 'win32') {
      execSync(`powershell -Command "Expand-Archive -Path '${archivePath}' -DestinationPath '${os.tmpdir()}' -Force"`)
    } else {
      execSync(`tar -xzf "${archivePath}" -C "${os.tmpdir()}" mln`)
    }

    fs.copyFileSync(path.join(os.tmpdir(), binaryName), binaryDest)

    if (process.platform !== 'win32') {
      fs.chmodSync(binaryDest, 0o755)
    }

    fs.unlinkSync(archivePath)
    console.log(`mln ${version}: installed successfully`)
  } catch (e) {
    console.error(`mln: extraction failed: ${e.message}`)
    process.exit(1)
  }
})
